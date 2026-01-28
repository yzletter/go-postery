package repository

import (
	"context"

	"github.com/yzletter/go-postery/lottery/model"
)

type OrderRepository interface {
	CreateTempOrder(ctx context.Context, uid, gid int64) error
	DeleteTempOrder(ctx context.Context, uid int64) error
	GetTempOrder(ctx context.Context, uid int64) (int64, error)
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
