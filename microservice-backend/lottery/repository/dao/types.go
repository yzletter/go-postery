package dao

import (
	"context"

	model2 "github.com/yzletter/go-postery/microservice-backend/lottery/model"
)

type OrderDAO interface {
	Create(ctx context.Context, order *model2.Order) error
	Get(ctx context.Context, uid int64) (*model2.Order, error)
}

type GiftDAO interface {
	GetAll(ctx context.Context) ([]*model2.Gift, error)
	GetByID(ctx context.Context, gid int64) (*model2.Gift, error)
}
