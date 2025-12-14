package dao

import (
	"github.com/yzletter/go-postery/model"
)

// 定义 DAO 层所有接口

type UserDAO interface {
	Create(user *model.User) (*model.User, error)
	Delete(id int64) error

	GetPasswordHash(id int64) (string, error)
	GetStatus(id int64) (uint8, error)
	GetByID(id int64) (*model.User, error)
	GetByUsername(username string) (*model.User, error)

	UpdatePasswordHash(id int64, newHash string) error
	UpdateProfile(id int64, updates map[string]any) error
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
