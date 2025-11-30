package repository

import (
	"github.com/yzletter/go-postery/model"
)

// UserRepository 定义需要实现的接口
type UserRepository interface {
	Create(name, password string) (int, error)
	Delete(uid int) (bool, error)
	UpdatePassword(uid int, oldPass, newPass string) (bool, error)
	GetByID(uid int) *model.User
	GetByName(name string) *model.User
}
