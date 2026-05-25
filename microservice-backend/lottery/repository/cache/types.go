package cache

import (
	"context"

	"github.com/yzletter/go-postery/microservice-backend/lottery/model"
)

type OrderCache interface {
	CreateTempOrder(ctx context.Context, order *model.TempOrder) error
	DeleteTempOrder(ctx context.Context, uid, tempOrderID int64) error
	GetTempOrder(ctx context.Context, uid int64) (*model.TempOrder, error)
}

type GiftCache interface {
	InitInventory(ctx context.Context, gifts []*model.Gift)
	GetAllInventory(ctx context.Context) ([]*model.Gift, error)
	ReduceInventory(ctx context.Context, gid int64) error
	IncreaseInventory(ctx context.Context, gid int64) error
}
