package service

import (
	"github.com/yzletter/go-postery/model"
	repository "github.com/yzletter/go-postery/repository/post"
)

type PostService struct {
	PostRepository *repository.GormPostRepository
}

func NewPostService(postRepository *repository.GormPostRepository) *PostService {
	return &PostService{PostRepository: postRepository}
}

func (svc *PostService) Create(uid int, title, content string) (int, error) {
	pid, err := svc.PostRepository.Create(uid, title, content)
	return pid, err
}

func (svc *PostService) Delete(pid int) error {
	err := svc.PostRepository.Delete(pid)
	return err
}
func (svc *PostService) Update(pid int, title, content string) error {
	err := svc.PostRepository.Update(pid, title, content)
	return err
}

func (svc *PostService) GetByPage(pageNo, pageSize int) (int, []*model.Post) {
	// 获取帖子总数和当前页帖子列表
	total, posts := svc.PostRepository.GetByPage(pageNo, pageSize)
	return total, posts
}

func (svc *PostService) HasMore(pageNo, pageSize, total int) bool {
	return pageNo*pageSize < total
}

func (svc *PostService) GetById(pid int) *model.Post {
	post := svc.PostRepository.GetByID(pid)
	return post
}
