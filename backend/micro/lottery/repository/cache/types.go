package cache

import (
	"context"

	"github.com/yzletter/go-postery/backend/micro/lottery/model"
)

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
