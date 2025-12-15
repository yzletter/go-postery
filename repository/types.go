package repository

import (
	"context"

	"github.com/yzletter/go-postery/model"
)

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
	Create(ctx context.Context, post *model.Post) (*model.Post, error)
	Delete(ctx context.Context, id int64) error
	UpdateCount(ctx context.Context, id int64, field model.PostCntField, delta int) error
	Update(ctx context.Context, id int64, updates map[string]any) error
	GetByID(ctx context.Context, id int64) (*model.Post, error)
	GetByUid(ctx context.Context, id int64, pageNo, pageSize int) (int64, []*model.Post, error)
	GetByPage(ctx context.Context, pageNo, pageSize int) (int64, []*model.Post, error)
	GetByPageAndTag(ctx context.Context, tid int64, pageNo, pageSize int) (int64, []*model.Post, error)
}
type CommentRepository interface {
}
type TagRepository interface {
}
type LikeRepository interface {
}
type FollowRepository interface {
}
