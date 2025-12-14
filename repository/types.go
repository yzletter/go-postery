package repository

import "github.com/yzletter/go-postery/model"

type UserRepository interface {
	Create(user *model.User) (*model.User, error)
	Delete(id int64) error
	GetPasswordHash(id int64) (string, error)
	GetStatus(id int64) (int, error)
	GetByID(id int64) (*model.User, error)
	GetByUsername(username string) (*model.User, error)
	UpdatePasswordHash(id int64, newHash string) error
	UpdateProfile(id int64, updates map[string]any) error
}

type PostRepository interface {
}
type CommentRepository interface {
}
type TagRepository interface {
}
type LikeRepository interface {
}
type FollowRepository interface {
}
