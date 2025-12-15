package dao

import (
	"context"

	"github.com/yzletter/go-postery/model"
)

// 定义 DAO 层所有接口

type UserDAO interface {
	Create(ctx context.Context, user *model.User) (*model.User, error)
	Delete(ctx context.Context, id int64) error
	GetPasswordHash(ctx context.Context, id int64) (string, error)
	GetStatus(ctx context.Context, id int64) (int, error)
	GetByID(ctx context.Context, id int64) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	UpdatePasswordHash(ctx context.Context, id int64, newHash string) error
	UpdateProfile(ctx context.Context, id int64, updates map[string]any) error
}

type PostDAO interface {
	Create(ctx context.Context, post *model.Post) (*model.Post, error)
	Delete(ctx context.Context, id int64) error
	UpdateCount(ctx context.Context, id int64, field model.PostCntField, delta int) error
	Update(ctx context.Context, id int64, updates map[string]any) error
	GetByID(ctx context.Context, id int64) (*model.Post, error)
	GetByUid(ctx context.Context, id int64, pageNo, pageSize int) (int64, []*model.Post, error)
	GetByPage(ctx context.Context, pageNo, pageSize int) (int64, []*model.Post, error)
	GetByPageAndTag(ctx context.Context, tid int64, pageNo, pageSize int) (int64, []*model.Post, error)
}

type CommentDAO interface {
	Create(ctx context.Context, comment *model.Comment) (*model.Comment, error)
	GetByID(ctx context.Context, id int64) (*model.Comment, error)
	Delete(ctx context.Context, id int64) (int, error)
	GetByPostID(ctx context.Context, id int64, pageNo, pageSize int) (int64, []*model.Comment, error)
	GetRepliesByParentID(ctx context.Context, id int64) ([]*model.Comment, error)
}

type LikeDAO interface {
}

type TagDAO interface{}

type FollowDAO interface {
	Follow(ctx context.Context, ferID, feeID int64) error
	UnFollow(ctx context.Context, ferID, feeID int64) error
	IfFollow(ctx context.Context, ferID, feeID int64) (int, error)
	GetFollowers(ctx context.Context, id int64, pageNo, pageSize int) (int64, []int64, error)
	GetFollowees(ctx context.Context, id int64, pageNo, pageSize int) (int64, []int64, error)
}
