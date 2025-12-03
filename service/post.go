package service

import (
	"errors"
	"log/slog"

	"github.com/yzletter/go-postery/model"
	repository "github.com/yzletter/go-postery/repository/post"
	userRepository "github.com/yzletter/go-postery/repository/user"
)

type PostService struct {
	PostRepository *repository.GormPostRepository
	UserRepository *userRepository.GormUserRepository
}

func NewPostService(postRepository *repository.GormPostRepository, userRepository *userRepository.GormUserRepository) *PostService {
	return &PostService{
		PostRepository: postRepository,
		UserRepository: userRepository,
	}
}

func (svc *PostService) Create(uid int, title, content string) (int, error) {
	pid, err := svc.PostRepository.Create(uid, title, content)
	return pid, err
}

func (svc *PostService) Delete(pid, uid int) error {
	// 判断登录用户是否是作者
	ok := svc.Belong(pid, uid)
	if !ok {
		// 无权限删除
		return errors.New("没有权限")
	}

	// 删除帖子
	err := svc.PostRepository.Delete(pid)
	return err
}
func (svc *PostService) Update(pid int, uid int, title, content string) error {
	// 判断登录用户是否是作者
	ok := svc.Belong(pid, uid)
	if !ok {
		// 无权限删除
		return errors.New("没有权限")
	}

	err := svc.PostRepository.Update(pid, title, content)
	return err
}

func (svc *PostService) GetByPage(pageNo, pageSize int) (int, []*model.Post) {
	// 获取帖子总数和当前页帖子列表
	total, posts := svc.PostRepository.GetByPage(pageNo, pageSize)
	for _, post := range posts {
		// 根据 uid 找到 username 进行赋值
		user := svc.UserRepository.GetByID(post.UserId)
		if user != nil {
			post.UserName = user.Name
		} else {
			slog.Warn("could not get name of user", "uid", post.UserId)
		}

		return total, posts
	}
	return 0, nil
}
func (svc *PostService) GetById(pid int) *model.Post {
	post := svc.PostRepository.GetByID(pid)
	return post
}

func (svc *PostService) HasMore(pageNo, pageSize, total int) bool {
	return pageNo*pageSize < total
}

// Belong 判断登录用户是否是帖子作者
func (svc *PostService) Belong(pid, uid int) bool {
	post := svc.GetById(pid)
	if post == nil || uid != post.UserId {
		return false
	}
	return true
}
