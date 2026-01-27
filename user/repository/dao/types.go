package dao

import (
	"context"

	"github.com/yzletter/go-postery/user/model"
)

type UserDAO interface {
	GetProfileByID(ctx context.Context, id int64) (*model.UserProfile, error)  // 根据 ID 查找用户资料
	UpdateProfile(ctx context.Context, id int64, updates map[string]any) error // 根据 ID 修改用户资料的多个字段
}
