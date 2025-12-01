package service

import repository "github.com/yzletter/go-postery/repository/comment"

type CommentService struct {
	CommentRepository *repository.CommentRepository
}

func NewCommentService(commentRepository *repository.CommentRepository) *CommentService {
	return &CommentService{
		CommentRepository: commentRepository,
	}
}

func (svc *CommentService) Create(pid int, uid int, parentId int, content string) (int, error) {
	cid, err := svc.CommentRepository.Create(pid, uid, parentId, content)
	return cid, err
}
