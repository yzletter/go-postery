package service

import (
	"github.com/yzletter/go-postery/model"
	repository "github.com/yzletter/go-postery/repository/post"
)

type PostService struct {
	PostRepository *repository.GormPostRepository
}

func (service *PostService) Create(uid int, title, content string) (int, error) {
	pid, err := service.PostRepository.Create(uid, title, content)
	return pid, err
}

func (service *PostService) Delete(pid int) error {
	err := service.PostRepository.Delete(pid)
	return err
}
func (service *PostService) Update(pid int, title, content string) error {
	err := service.PostRepository.Update(pid, title, content)
	return err
}

func NewPostService(postRepository *repository.GormPostRepository) *PostService {
	return &PostService{PostRepository: postRepository}
}

func (service *PostService) GetByPage(pageNo, pageSize int) (int, []*model.Post) {
	// 获取帖子总数和当前页帖子列表
	total, posts := service.PostRepository.GetByPage(pageNo, pageSize)
	return total, posts
}

func (service *PostService) HasMore(pageNo, pageSize, total int) bool {
	return pageNo*pageSize < total
}

func (service *PostService) GetById(pid int) *model.Post {
	post := service.PostRepository.GetByID(pid)
	return post
}
