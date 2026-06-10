package repository

import (
	"context"
	"time"

	model2 "github.com/yzletter/go-postery/micro-backend/lottery/model"
)

type OrderRepository interface {
	CreateTempOrder(ctx context.Context, order *model2.Order) error               // 创建临时订单
	PayTempOrder(ctx context.Context, orderID int64) error                        // 支付订单
	CancelTempOrder(ctx context.Context, orderID int64) error                     // 取消临时订单
	RecycleTempOrder(ctx context.Context, uid int64, orderID int64) (bool, error) // 回收临时订单，返回是否需要回补库存
	MarkRollbackDone(ctx context.Context, orderID int64) error
	MarkRollbackFailed(ctx context.Context, orderID int64, nextRollbackAt time.Time) error
	ListRollbackDueOrders(ctx context.Context, limit int) ([]*model2.Order, error)
	GetTempOrder(ctx context.Context, uid int64) (*model2.Order, error)
	GetOrder(ctx context.Context, uid int64) (*model2.Order, error)
}

type GiftRepository interface {
	GetAllGifts(ctx context.Context) ([]*model2.Gift, error)
	GetCacheInventory(ctx context.Context) ([]*model2.Gift, error)
	GetByID(ctx context.Context, gid int64) (*model2.Gift, error)
	ReduceCacheInventory(ctx context.Context, gid int64) error
	IncreaseCacheInventory(ctx context.Context, gid int64) error
	RollbackCacheInventory(ctx context.Context, orderID, gid int64) error
	InitCacheInventory(ctx context.Context)
}
