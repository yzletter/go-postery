package repository

import (
	"github.com/yzletter/go-postery/model"
)

// UserRepository 定义接口
type UserRepository interface {
	Create(uid int) error
	Delete(uid int) (bool, error)
	Update(user *model.User) (bool, error)
	GetByID(userID int) (model.User, error)
	GetByName(userName string) (model.User, error)
}
