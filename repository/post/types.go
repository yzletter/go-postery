package repository

import "github.com/yzletter/go-postery/model"

// PostRepository 定义接口 todo 复制的需要改
type PostRepository interface {
	Create(uid int) error
	Delete(uid int) (bool, error)
	Update(user *model.User) (bool, error)
	GetByID(userID int) (model.User, error)
}
