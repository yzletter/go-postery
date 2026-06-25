package cache

import (
	"context"

	"github.com/yzletter/go-postery/backend/micro/lottery/dto"
	"github.com/yzletter/go-postery/backend/micro/lottery/model"
)

type OrderCache interface {
	// CreateTempOrder 创建临时订单缓存
	//
	// Parameter:
	//	- order: 订单
	//
	// Return:
	//	- error: 可能返回的错误
	CreateTempOrder(ctx context.Context, order *dto.Order) error

	// DeleteTempOrder 删除临时订单缓存
	//
	// Parameter:
	//	- uid: 用户 ID
	//	- tempOrderID: 临时订单 ID
	//
	// Return:
	//	- error: 可能返回的错误
	DeleteTempOrder(ctx context.Context, uid, tempOrderID int64) error

	// RecycleTempOrder 回收临时订单缓存
	//
	// Parameter:
	//	- uid: 用户 ID
	//	- tempOrderID: 临时订单 ID
	//
	// Return:
	//	- bool: 是否需要回补库存
	//	- error: 可能返回的错误
	RecycleTempOrder(ctx context.Context, uid, tempOrderID int64) (bool, error)

	// GetTempOrder 获取临时订单缓存
	//
	// Parameter:
	//	- uid: 用户 ID
	//
	// Return:
	//	- *dto.Order: 临时订单
	//	- error: 可能返回的错误
	GetTempOrder(ctx context.Context, uid int64) (*dto.Order, error)
}

type GiftCache interface {
	// InitInventory 初始化库存缓存
	//
	// Parameter:
	//	- gifts: 奖品列表
	InitInventory(ctx context.Context, gifts []*model.Gift)

	// GetAllInventory 获取所有库存缓存
	//
	// Return:
	//	- []*model.Gift: 奖品库存列表
	//	- error: 可能返回的错误
	GetAllInventory(ctx context.Context) ([]*model.Gift, error)

	// ReduceInventory 减少库存缓存
	//
	// Parameter:
	//	- gid: 奖品 ID
	//
	// Return:
	//	- error: 可能返回的错误
	ReduceInventory(ctx context.Context, gid int64) error

	// IncreaseInventory 增加库存缓存
	//
	// Parameter:
	//	- gid: 奖品 ID
	//
	// Return:
	//	- error: 可能返回的错误
	IncreaseInventory(ctx context.Context, gid int64) error

	// RollbackInventory 回滚库存缓存
	//
	// Parameter:
	//	- orderID: 订单 ID
	//	- gid: 奖品 ID
	//
	// Return:
	//	- error: 可能返回的错误
	RollbackInventory(ctx context.Context, orderID, gid int64) error
}
