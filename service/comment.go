package service

import (
	"errors"
	"log/slog"

	"github.com/yzletter/go-postery/infra/model"
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

func (svc *CommentService) Create(pid int, uid int, parentId int, content string) (int, error) {
	cid, err := svc.CommentRepository.Create(pid, uid, parentId, content)
	return cid, err
}

func (svc *CommentService) Delete(uid int, cid int) error {
	comment := svc.CommentRepository.GetByID(cid)
	if comment == nil {
		slog.Error("评论不存在", "cid", cid)
		return errors.New("评论不存在")
	}

	if comment.UserId != uid {
		return errors.New("没有删除权限")
	}

	err := svc.CommentRepository.Delete(cid)
	if err != nil {
		return errors.New("删除失败")
	}
	return nil
}

func (svc *CommentService) List(pid int) []*model.Comment {
	comments := svc.CommentRepository.GetByPostID(pid)
	if comments == nil {
		return nil
	}

	for _, comment := range comments {
		name := svc.UserRepository.GetByID(comment.UserId).Name
		comment.UserName = name
	}

	return comments
}
