package dao

import (
	"github.com/yzletter/go-postery/model"
)

// 定义 DAO 层所有接口

type UserDAO interface {
	Create(name, password string) (model.User, error)
	Delete(id int) error
	UpdatePassword(id int, oldPass, newPass string) error
	UpdateProfile(id int, request model.User) error
	GetByID(id int) (model.User, error)
	GetByName(name string) (model.User, error)
	Status(id int) (int, error)
}

type PostDAO interface {
}

type CommentDAO interface {
}

type LikeDAO interface {
}

type TagDAO interface {
}

type FollowDAO interface {
}
