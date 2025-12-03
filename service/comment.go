package service

import (
	"errors"
	"log/slog"

	dto "github.com/yzletter/go-postery/dto/response"
	repository "github.com/yzletter/go-postery/repository/comment"
	userRepository "github.com/yzletter/go-postery/repository/user"
)

type CommentService struct {
	CommentRepository *repository.GormCommentRepository
	UserRepository    *userRepository.GormUserRepository
}

func NewCommentService(commentRepository *repository.GormCommentRepository, userRepository *userRepository.GormUserRepository) *CommentService {
	return &CommentService{
		CommentRepository: commentRepository,
		UserRepository:    userRepository,
	}
}

func (svc *CommentService) Create(pid int, uid int, parentId int, content string) (dto.CommentDTO, error) {
	comment, err := svc.CommentRepository.Create(pid, uid, parentId, content)
	_, user := svc.UserRepository.GetByID(uid)
	return dto.ToCommentDTO(comment, user), err
}

func (svc *CommentService) Delete(uid int, cid int) error {
	comment, err := svc.CommentRepository.GetByID(cid)
	if err != nil {
		slog.Error("评论不存在", "cid", cid)
		return errors.New("评论不存在")
	}

	if comment.UserId != uid {
		return errors.New("没有删除权限")
	}

	err = svc.CommentRepository.Delete(cid)
	if err != nil {
		return errors.New("删除失败")
	}
	return nil
}

func (svc *CommentService) List(pid int) []dto.CommentDTO {
	comments := svc.CommentRepository.GetByPostID(pid)
	if comments == nil {
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
	return comment.UserId == uid
}
