package dao

import (
	"context"

	"github.com/yzletter/go-postery/microservice-backend/lottery/model"
)

type OrderDAO interface {
	Create(ctx context.Context, order *model.Order) error
	CreateTempOrder(ctx context.Context, tempOrder *model.Order) error
	Get(ctx context.Context, uid int64) (*model.Order, error)
}

type GiftDAO interface {
	GetAll(ctx context.Context) ([]*model.Gift, error)
	GetByID(ctx context.Context, gid int64) (*model.Gift, error)
}
