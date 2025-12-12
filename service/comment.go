package service

import (
	"errors"

	dto "github.com/yzletter/go-postery/dto/response"
	commentRepository "github.com/yzletter/go-postery/repository/comment"
	postRepository "github.com/yzletter/go-postery/repository/post"
	userRepository "github.com/yzletter/go-postery/repository/user"
)

var fields = []string{"view_count", "comment_count", "like_count"}

type CommentService struct {
	CommentDBRepo    *commentRepository.CommentDBRepository
	CommentCacheRepo *commentRepository.CommentCacheRepository
	UserDBRepo       *userRepository.UserDBRepository
	PostDBRepo       *postRepository.PostDBRepository
	PostCacheRepo    *postRepository.PostCacheRepository
}

func NewCommentService(commentDBRepo *commentRepository.CommentDBRepository, commentCacheRepo *commentRepository.CommentCacheRepository,
	userDBRepo *userRepository.UserDBRepository,
	postDBRepo *postRepository.PostDBRepository, postCacheRepo *postRepository.PostCacheRepository) *CommentService {
	return &CommentService{
		CommentDBRepo:    commentDBRepo,
		CommentCacheRepo: commentCacheRepo,
		UserDBRepo:       userDBRepo,
		PostDBRepo:       postDBRepo,
		PostCacheRepo:    postCacheRepo,
	}
}

func (svc *CommentService) Create(pid int, uid int, parentId int, replyId int, content string) (dto.CommentDTO, error) {
	comment, err := svc.CommentDBRepo.Create(pid, uid, parentId, replyId, content)
	_, user := svc.UserDBRepo.GetByID(uid)

	svc.PostDBRepo.ChangeCommentCnt(pid, 1)
	ok, err := svc.PostCacheRepo.ChangeCommentCnt(pid, 1)
	if !ok {
		ok, post := svc.PostDBRepo.GetByID(pid)
		if ok {
			vals := []int{post.ViewCount, post.CommentCount, post.LikeCount}
			svc.PostCacheRepo.SetKey(pid, fields, vals)
		}
	}

	return dto.ToCommentDTO(comment, user), err
}

func (svc *CommentService) Delete(uid, pid, cid int) error {
	ok := svc.Belong(cid, uid)
	if !ok {
		return errors.New("没有删除权限")
	}

	ok, post := svc.PostDBRepo.GetByID(pid)

	cnt, err := svc.CommentDBRepo.Delete(cid) // 返回被删除的个数
	if err != nil {
		return errors.New("删除失败")
	}

	svc.PostDBRepo.ChangeCommentCnt(pid, -cnt)
	ok, err = svc.PostCacheRepo.ChangeCommentCnt(pid, -cnt)
	if !ok {
		vals := []int{post.ViewCount, post.CommentCount - cnt, post.LikeCount}
		svc.PostCacheRepo.SetKey(pid, fields, vals)
	}

	return nil
}

func (svc *CommentService) List(pid int) []dto.CommentDTO {
	comments, err := svc.CommentDBRepo.GetByPostID(pid)
	if err != nil {
		return nil
	}

	var commentDTOs []dto.CommentDTO
	for _, comment := range comments {
		_, user := svc.UserDBRepo.GetByID(comment.UserId)
		commentDTO := dto.ToCommentDTO(comment, user)
		commentDTOs = append(commentDTOs, commentDTO)
	}

	return commentDTOs
}

func (svc *CommentService) Belong(cid, uid int) bool {
	comment, err := svc.CommentDBRepo.GetByID(cid)
	if err != nil {
		return false
	}

	ok, post := svc.PostDBRepo.GetByID(comment.PostId)
	if !ok {
		return false
	}

	// 帖子属于当前登录用户，或评论属于当前用户
	return comment.UserId == uid || post.UserId == uid
}
