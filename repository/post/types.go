package repository

import "github.com/yzletter/go-postery/model"

// PostRepository 定义接口
type PostRepository interface {
	Create(uid int, title, content string) (int, error)
	Delete(pid int) error
	Update(pid int, title, content string) error
	GetByID(pid int) *model.Post
	GetByPage(pageNo, pageSize int) (int, []*model.Post)
	GetByUid(uid int) []*model.Post
}
