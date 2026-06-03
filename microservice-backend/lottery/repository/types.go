package repository

import (
	"context"

	"github.com/yzletter/go-postery/microservice-backend/lottery/model"
)

type OrderRepository interface {
	CreateTempOrder(ctx context.Context, order *model.Order) error                // 创建临时订单
	PayTempOrder(ctx context.Context, orderID int64) error                        // 支付订单
	CancelTempOrder(ctx context.Context, orderID int64) error                     // 取消临时订单
	RecycleTempOrder(ctx context.Context, uid int64, orderID int64) (bool, error) // 回收临时订单
	GetTempOrder(ctx context.Context, uid int64) (*model.Order, error)
	GetOrder(ctx context.Context, uid int64) (*model.Order, error)
}

type GiftRepository interface {
	GetAllGifts(ctx context.Context) ([]*model.Gift, error)
	GetCacheInventory(ctx context.Context) ([]*model.Gift, error)
	GetByID(ctx context.Context, gid int64) (*model.Gift, error)
	ReduceCacheInventory(ctx context.Context, gid int64) error
	IncreaseCacheInventory(ctx context.Context, gid int64) error
	InitCacheInventory(ctx context.Context)
}
