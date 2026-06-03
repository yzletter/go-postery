package repository

import (
	"context"

	"github.com/yzletter/go-postery/microservice-backend/lottery/model"
)

type OrderRepository interface {
	CreateTempOrder(ctx context.Context, order *model.Order) error
	PayTempOrder(ctx context.Context, orderID int64) error
	CancelTempOrder(ctx context.Context, orderID int64) error
	DeleteTempOrder(ctx context.Context, uid, tempOrderID int64) error
	RecycleTempOrder(ctx context.Context, uid int64, tempOrderID int64) (bool, error)
	GetTempOrder(ctx context.Context, uid int64) (*model.Order, error)
	CreateOrder(ctx context.Context, order *model.Order) error
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
