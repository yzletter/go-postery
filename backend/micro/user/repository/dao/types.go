package dao

import (
	"context"
	"time"

	"github.com/yzletter/go-postery/backend/micro/user/model"
)

// UserDAO 定义用户资料 DAO 接口
type UserDAO interface {
	// GetProfile 根据 ID 查找用户资料
	//
	// Parameter:
	//	- id: 用户 ID
	//
	// Return:
	//	- *model.Profile: 用户资料
	//	- error: 可能返回的错误
	//		- ErrServerInternal: 数据库内部错误
	//		- ErrRecordNotFound: 用户资料不存在
	GetProfile(ctx context.Context, id int64) (*model.Profile, error)

	// GetIDAfterTime 根据时间查找之后创建的用户 ID
	//
	// Parameter:
	//	- timeAfter: 查询起始时间
	//
	// Return:
	//	- []int64: 用户 ID 列表
	//	- error: 可能返回的错误
	//		- ErrServerInternal: 数据库内部错误
	GetIDAfterTime(ctx context.Context, timeAfter time.Time) ([]int64, error)

	// UpdateProfile 根据 ID 修改用户资料的多个字段
	//
	// Parameter:
	//	- id: 用户 ID
	//	- updates: 待更新字段
	//
	// Return:
	//	- error: 可能返回的错误
	//		- ErrServerInternal: 数据库内部错误
	//		- ErrRecordNotFound: 用户资料不存在
	//		- ErrUniqueKey: 唯一键冲突
	UpdateProfile(ctx context.Context, id int64, updates map[string]any) error
}
