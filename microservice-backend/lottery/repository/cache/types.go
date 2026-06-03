package cache

import (
	"context"

	"github.com/yzletter/go-postery/microservice-backend/lottery/dto"
	"github.com/yzletter/go-postery/microservice-backend/lottery/model"
)

type OrderCache interface {
	CreateTempOrder(ctx context.Context, order *dto.Order) error
	DeleteTempOrder(ctx context.Context, uid, tempOrderID int64) error
	RecycleTempOrder(ctx context.Context, uid, tempOrderID int64) (bool, error)
	GetTempOrder(ctx context.Context, uid int64) (*dto.Order, error)
}

type GiftCache interface {
	InitInventory(ctx context.Context, gifts []*model.Gift)
	GetAllInventory(ctx context.Context) ([]*model.Gift, error)
	ReduceInventory(ctx context.Context, gid int64) error
	IncreaseInventory(ctx context.Context, gid int64) error
}
