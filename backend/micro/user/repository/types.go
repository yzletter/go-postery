package repository

import (
	"context"
	"time"

	"github.com/yzletter/go-postery/backend/micro/user/domain"
)

// UserRepository 定义用户资料仓储接口
type UserRepository interface {
	// GetProfile 根据 ID 查找用户资料
	//
	// Parameter:
	//	- uid: 用户 ID
	//
	// Return:
	//	- domain.Profile: 用户资料
	//	- error: 可能返回的错误
	//		- ErrServerInternal: 系统内部错误
	//		- ErrRecordNotFound: 用户资料不存在
	GetProfile(ctx context.Context, uid int64) (domain.Profile, error)

	// GetIDAfterTime 根据时间查找之后创建的用户 ID
	//
	// Parameter:
	//	- timeAfter: 查询起始时间
	//
	// Return:
	//	- []int64: 用户 ID 列表
	//	- error: 可能返回的错误
	//		- ErrServerInternal: 系统内部错误
	GetIDAfterTime(ctx context.Context, timeAfter time.Time) ([]int64, error)

	// UpdateProfile 根据 ID 修改用户资料的多个字段
	//
	// Parameter:
	//	- id: 用户 ID
	//	- updates: 待更新字段
	//
	// Return:
	//	- error: 可能返回的错误
	//		- ErrServerInternal: 系统内部错误
	//		- ErrRecordNotFound: 用户资料不存在
	//		- ErrUniqueKey: 唯一键冲突
	UpdateProfile(ctx context.Context, id int64, updates map[string]any) error
}
