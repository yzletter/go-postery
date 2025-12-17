package service

import (
	"context"
	"errors"
	"log/slog"

	commentdto "github.com/yzletter/go-postery/dto/comment"
	"github.com/yzletter/go-postery/errno"
	"github.com/yzletter/go-postery/model"
	"github.com/yzletter/go-postery/repository"
)

type commentService struct {
	CommentRepo repository.CommentRepository
	UserRepo    repository.UserRepository
	PostRepo    repository.PostRepository
	idGen       IDGenerator
}

func NewCommentService(commentRepo repository.CommentRepository, postRepo repository.PostRepository, userRepo repository.UserRepository, idGen IDGenerator) CommentService {
	return &commentService{
		CommentRepo: commentRepo,
		PostRepo:    postRepo,
		UserRepo:    userRepo,
		idGen:       idGen,
	}
}

func (svc *commentService) Create(ctx context.Context, pid int64, uid int64, parentId int64, replyId int64, content string) (commentdto.DTO, error) {
	var empty commentdto.DTO

	// 查询作者
	author, err := svc.UserRepo.GetByID(ctx, uid)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return empty, errno.ErrUserNotFound
		}
		return empty, errno.ErrServerInternal
	}

	// 查询帖子
	_, err = svc.PostRepo.GetByID(ctx, pid)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return empty, errno.ErrPostNotFound
		}
		return empty, errno.ErrServerInternal
	}

	// 新建评论
	comment := &model.Comment{
		ID:       svc.idGen.NextID(),
		PostID:   pid,
		ParentID: parentId,
		ReplyID:  replyId,
		UserID:   uid,
		Content:  content,
	}
	err = svc.CommentRepo.Create(ctx, comment)
	if err != nil {
		if errors.Is(err, repository.ErrUniqueKey) {
			// 雪花 ID 的评论不会已存在, 需要排查
			slog.Error("Create Post Failed", "error", err)
		}
		return empty, errno.ErrServerInternal
	}

	// 修改评论数
	field := model.PostCommentCount
	err = svc.PostRepo.UpdateCount(ctx, pid, field, 1)
	if err != nil {
		slog.Error("Update Comment Count Failed", "error", err)
	}

	return commentdto.ToDTO(comment, author), err
}

func (svc *commentService) Delete(ctx context.Context, uid, cid int64) error {
	// 判断是否有删除权限
	ok := svc.CheckAuth(ctx, cid, uid)
	if !ok {
		return errno.ErrUnauthorized
	}

	comment, err := svc.CommentRepo.GetByID(ctx, cid)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return errno.ErrCommentNotFound
		}
		return errno.ErrServerInternal
	}

	// 删除评论
	cnt, err := svc.CommentRepo.Delete(ctx, cid) // 返回被删除的个数
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return errno.ErrCommentNotFound
		}
		return errno.ErrServerInternal
	}

	// 改变评论数
	field := model.PostCommentCount
	err = svc.PostRepo.UpdateCount(ctx, comment.PostID, field, -cnt)
	if err != nil {
		slog.Error("Update Comment Failed", "error", err)
	}

	return nil
}

func (svc *commentService) List(ctx context.Context, pid int64, pageNo, pageSize int) (int, []commentdto.DTO, error) {
	var empty []commentdto.DTO
	total, comments, err := svc.CommentRepo.GetByPostID(ctx, pid, pageNo, pageSize)
	if err != nil {
		return 0, empty, errno.ErrCommentNotFound
	}

	var commentDTOs []commentdto.DTO
	for _, comment := range comments {
		user, err := svc.UserRepo.GetByID(ctx, comment.UserID)
		if err != nil {
			user = &model.User{}
		}
		commentDTO := commentdto.ToDTO(comment, user)
		commentDTOs = append(commentDTOs, commentDTO)
	}

	return int(total), commentDTOs, nil
}

// CheckAuth 判断是否有删除权限
func (svc *commentService) CheckAuth(ctx context.Context, cid, uid int64) bool {
	comment, err := svc.CommentRepo.GetByID(ctx, cid)
	if err != nil {
		return false
	}

	post, err := svc.PostRepo.GetByID(ctx, comment.PostID)
	if err != nil {
		return false
	}

	// 帖子属于当前登录用户，或评论属于当前用户
	return comment.UserID == uid || post.UserID == uid
}
