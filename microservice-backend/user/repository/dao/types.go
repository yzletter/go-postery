package dao

import (
	"context"

	model2 "github.com/yzletter/go-postery/microservice-backend/user/model"
)

type UserDAO interface {
	GetProfileByID(ctx context.Context, id int64) (*model2.UserProfile, error) // 根据 ID 查找用户资料
	UpdateProfile(ctx context.Context, id int64, updates map[string]any) error // 根据 ID 修改用户资料的多个字段
	UpdateAvatar(ctx context.Context, uid int64, avatar string) error          // 修改用户头像链接
}

type FollowDAO interface {
	Create(ctx context.Context, follow *model2.Follow) error
	Delete(ctx context.Context, ferID, feeID int64) error
	Exists(ctx context.Context, ferID, feeID int64) (model2.FollowType, error)
	GetFollowers(ctx context.Context, id int64, pageNo, pageSize int) (int64, []int64, error)
	GetFollowees(ctx context.Context, id int64, pageNo, pageSize int) (int64, []int64, error)
}
