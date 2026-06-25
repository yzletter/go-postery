package cache

import (
	"context"

	"github.com/yzletter/go-postery/backend/micro/user/domain"
)

// UserCache 定义用户资料缓存接口
type UserCache interface {
	// GetProfile 获取用户资料缓存
	//
	// Parameter:
	//	- id: 用户 ID
	//
	// Return:
	//	- domain.Profile: 用户资料
	//	- error: 可能返回的错误
	//		- redis.Nil: 缓存不存在
	//		- ErrServerInternal: 反序列化失败
	//		- Redis 原始错误: 缓存读取失败
	GetProfile(ctx context.Context, id int64) (domain.Profile, error)

	// SetProfile 设置用户资料缓存
	//
	// Parameter:
	//	- id: 用户 ID
	//	- profile: 用户资料
	//
	// Return:
	//	- error: 可能返回的错误
	//		- ErrServerInternal: 序列化失败或缓存写入失败
	SetProfile(ctx context.Context, id int64, profile domain.Profile) error

	// DelProfile 删除用户资料缓存
	//
	// Parameter:
	//	- id: 用户 ID
	//
	// Return:
	//	- error: 可能返回的错误
	//		- Redis 原始错误: 缓存删除失败
	DelProfile(ctx context.Context, id int64) error
}
