package service

import (
	"errors"
	"log/slog"

	dto "github.com/yzletter/go-postery/dto/response"
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

func (svc *PostService) Create(uid int, title, content string) (dto.PostDTO, error) {
	post, err := svc.PostRepository.Create(uid, title, content)
	_, user := svc.UserRepository.GetByID(uid)
	postDTO := dto.ToPostDTO(post, user)
	return postDTO, err
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

func (svc *PostService) GetByPage(pageNo, pageSize int) (int, []dto.PostDTO) {
	// 获取帖子总数和当前页帖子列表
	total, posts := svc.PostRepository.GetByPage(pageNo, pageSize)

	var postDTOs []dto.PostDTO
	for _, post := range posts {
		// 根据 uid 找到 username 进行赋值
		ok, user := svc.UserRepository.GetByID(post.UserId)
		if !ok {
			slog.Warn("could not get name of user", "uid", post.UserId)
		}

		postDTO := dto.ToPostDTO(post, user)
		postDTOs = append(postDTOs, postDTO)
	}
	return total, postDTOs
}

func (svc *PostService) GetById(pid int) (bool, dto.PostDTO) {
	ok, post := svc.PostRepository.GetByID(pid)
	if !ok {
		return false, dto.PostDTO{}
	}

	// 查找作者信息
	_, user := svc.UserRepository.GetByID(post.UserId)

	postDTO := dto.ToPostDTO(post, user)
	return true, postDTO
}

func (svc *PostService) HasMore(pageNo, pageSize, total int) bool {
	return pageNo*pageSize < total
}

// Belong 判断登录用户是否是帖子作者
func (svc *PostService) Belong(pid, uid int) bool {
	ok, postDTO := svc.GetById(pid)
	if !ok || uid != postDTO.Author.Id {
		return false
	}
	return true
}

func (svc *PostService) GetByUid(uid int) []dto.PostDTO {
	posts := svc.PostRepository.GetByUid(uid)
	if posts == nil {
		return nil
	}

	postDTOs := make([]dto.PostDTO, 0, len(posts))
	for _, post := range posts {
		// 查找作者信息
		_, user := svc.UserRepository.GetByID(post.UserId)

		// 转成 DTO 返回给 Handler
		postDTO := dto.ToPostDTO(post, user)
		postDTOs = append(postDTOs, postDTO)
	}

	return postDTOs
}
