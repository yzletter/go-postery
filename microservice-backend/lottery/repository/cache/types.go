package cache

import (
	"context"

	"github.com/yzletter/go-postery/microservice-backend/lottery/model"
)

type OrderCache interface {
	CreateTempOrder(ctx context.Context, uid, gid int64) error
	DeleteTempOrder(ctx context.Context, uid int64) error
	GetTempOrderID(ctx context.Context, uid int64) (int64, error)
}

type GiftCache interface {
	InitInventory(ctx context.Context, gifts []*model.Gift)
	GetAllInventory(ctx context.Context) ([]*model.Gift, error)
	ReduceInventory(ctx context.Context, gid int64) error
	IncreaseInventory(ctx context.Context, gid int64) error
}
