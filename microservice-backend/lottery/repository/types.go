package repository

import (
	"context"

	model2 "github.com/yzletter/go-postery/microservice-backend/lottery/model"
)

type OrderRepository interface {
	CreateTempOrder(ctx context.Context, uid, gid int64) error
	DeleteTempOrder(ctx context.Context, uid int64) error
	GetTempOrder(ctx context.Context, uid int64) (int64, error)
	CreateOrder(ctx context.Context, order *model2.Order) error
	GetOrder(ctx context.Context, uid int64) (*model2.Order, error)
}

type GiftRepository interface {
	GetAllGifts(ctx context.Context) ([]*model2.Gift, error)
	GetCacheInventory(ctx context.Context) ([]*model2.Gift, error)
	GetByID(ctx context.Context, gid int64) (*model2.Gift, error)
	ReduceCacheInventory(ctx context.Context, gid int64) error
	IncreaseCacheInventory(ctx context.Context, gid int64) error
	InitCacheInventory(ctx context.Context)
}
