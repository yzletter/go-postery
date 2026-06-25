package repository

import (
	"context"
	"time"

	"github.com/yzletter/go-postery/backend/micro/lottery/model"
)

type OrderRepository interface {
	// CreateTempOrder 创建临时订单
	//
	// Parameter:
	//	- order: 订单
	//
	// Return:
	//	- error: 可能返回的错误
	CreateTempOrder(ctx context.Context, order *model.Order) error

	// PayTempOrder 支付订单
	//
	// Parameter:
	//	- orderID: 订单 ID
	//
	// Return:
	//	- error: 可能返回的错误
	PayTempOrder(ctx context.Context, orderID int64) error

	// CancelTempOrder 取消临时订单
	//
	// Parameter:
	//	- orderID: 订单 ID
	//
	// Return:
	//	- error: 可能返回的错误
	CancelTempOrder(ctx context.Context, orderID int64) error

	// RecycleTempOrder 回收临时订单
	//
	// Parameter:
	//	- uid: 用户 ID
	//	- orderID: 订单 ID
	//
	// Return:
	//	- bool: 是否需要回补库存
	//	- error: 可能返回的错误
	RecycleTempOrder(ctx context.Context, uid int64, orderID int64) (bool, error)

	// MarkRollbackDone 标记库存回滚完成
	//
	// Parameter:
	//	- orderID: 订单 ID
	//
	// Return:
	//	- error: 可能返回的错误
	MarkRollbackDone(ctx context.Context, orderID int64) error

	// MarkRollbackFailed 标记库存回滚失败
	//
	// Parameter:
	//	- orderID: 订单 ID
	//	- nextRollbackAt: 下次回滚时间
	//
	// Return:
	//	- error: 可能返回的错误
	MarkRollbackFailed(ctx context.Context, orderID int64, nextRollbackAt time.Time) error

	// ListRollbackDueOrders 获取到期回滚订单
	//
	// Parameter:
	//	- limit: 查询数量
	//
	// Return:
	//	- []*model.Order: 订单列表
	//	- error: 可能返回的错误
	ListRollbackDueOrders(ctx context.Context, limit int) ([]*model.Order, error)

	// GetTempOrder 获取临时订单
	//
	// Parameter:
	//	- uid: 用户 ID
	//
	// Return:
	//	- *model.Order: 临时订单
	//	- error: 可能返回的错误
	GetTempOrder(ctx context.Context, uid int64) (*model.Order, error)

	// GetOrder 获取订单
	//
	// Parameter:
	//	- uid: 用户 ID
	//
	// Return:
	//	- *model.Order: 订单
	//	- error: 可能返回的错误
	GetOrder(ctx context.Context, uid int64) (*model.Order, error)
}

type GiftRepository interface {
	// GetAllGifts 获取所有奖品
	//
	// Return:
	//	- []*model.Gift: 奖品列表
	//	- error: 可能返回的错误
	GetAllGifts(ctx context.Context) ([]*model.Gift, error)

	// GetCacheInventory 获取缓存库存
	//
	// Return:
	//	- []*model.Gift: 奖品库存列表
	//	- error: 可能返回的错误
	GetCacheInventory(ctx context.Context) ([]*model.Gift, error)

	// GetByID 根据 ID 获取奖品
	//
	// Parameter:
	//	- gid: 奖品 ID
	//
	// Return:
	//	- *model.Gift: 奖品
	//	- error: 可能返回的错误
	GetByID(ctx context.Context, gid int64) (*model.Gift, error)

	// ReduceCacheInventory 减少缓存库存
	//
	// Parameter:
	//	- gid: 奖品 ID
	//
	// Return:
	//	- error: 可能返回的错误
	ReduceCacheInventory(ctx context.Context, gid int64) error

	// IncreaseCacheInventory 增加缓存库存
	//
	// Parameter:
	//	- gid: 奖品 ID
	//
	// Return:
	//	- error: 可能返回的错误
	IncreaseCacheInventory(ctx context.Context, gid int64) error

	// RollbackCacheInventory 回滚缓存库存
	//
	// Parameter:
	//	- orderID: 订单 ID
	//	- gid: 奖品 ID
	//
	// Return:
	//	- error: 可能返回的错误
	RollbackCacheInventory(ctx context.Context, orderID, gid int64) error

	// InitCacheInventory 初始化缓存库存
	//
	// Parameter:
	//	- ctx: 上下文
	InitCacheInventory(ctx context.Context)
}
