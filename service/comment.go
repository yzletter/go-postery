package service

import (
	"errors"
	"log/slog"

	repository "github.com/yzletter/go-postery/repository/comment"
)

type CommentService struct {
	CommentRepository *repository.GormCommentRepository
}

func NewCommentService(commentRepository *repository.GormCommentRepository) *CommentService {
	return &CommentService{
		CommentRepository: commentRepository,
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
