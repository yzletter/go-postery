package service

import (
	"errors"

	dto "github.com/yzletter/go-postery/dto/response"
	commentRepository "github.com/yzletter/go-postery/repository/comment"
	postRepository "github.com/yzletter/go-postery/repository/post"
	userRepository "github.com/yzletter/go-postery/repository/user"
)

type CommentService struct {
	CommentRepository *commentRepository.GormCommentRepository
	UserRepository    *userRepository.GormUserRepository
	PostRepository    *postRepository.GormPostRepository
}

func NewCommentService(commentRepository *commentRepository.GormCommentRepository, userRepository *userRepository.GormUserRepository, postRepository *postRepository.GormPostRepository) *CommentService {
	return &CommentService{
		CommentRepository: commentRepository,
		UserRepository:    userRepository,
		PostRepository:    postRepository,
	}
}

func (svc *CommentService) Create(pid int, uid int, parentId int, replyId int, content string) (dto.CommentDTO, error) {
	comment, err := svc.CommentRepository.Create(pid, uid, parentId, replyId, content)
	_, user := svc.UserRepository.GetByID(uid)
	return dto.ToCommentDTO(comment, user), err
}

func (svc *CommentService) Delete(uid int, cid int) error {
	ok := svc.Belong(cid, uid)
	if !ok {
		return errors.New("删除失败")
	}

	err := svc.CommentRepository.Delete(cid)
	if err != nil {
		return errors.New("删除失败")
	}

	return nil
}

func (svc *CommentService) List(pid int) []dto.CommentDTO {
	comments, err := svc.CommentRepository.GetByPostID(pid)
	if err != nil {
		return nil
	}

	var commentDTOs []dto.CommentDTO
	for _, comment := range comments {
		_, user := svc.UserRepository.GetByID(comment.UserId)
		commentDTO := dto.ToCommentDTO(comment, user)
		commentDTOs = append(commentDTOs, commentDTO)
	}

	return commentDTOs
}

func (svc *CommentService) Belong(cid, uid int) bool {
	comment, err := svc.CommentRepository.GetByID(cid)
	if err != nil {
		return false
	}

	ok, post := svc.PostRepository.GetByID(comment.PostId)
	if !ok {
		return false
	}

	// 帖子属于当前登录用户，或评论属于当前用户
	return comment.UserId == uid || post.UserId == uid
}
