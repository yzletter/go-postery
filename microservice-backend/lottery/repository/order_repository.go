package repository

import (
	"context"

	"github.com/yzletter/go-postery/microservice-backend/lottery/model"
	"github.com/yzletter/go-postery/microservice-backend/lottery/repository/cache"
	"github.com/yzletter/go-postery/microservice-backend/lottery/repository/dao"
)

type orderRepository struct {
	dao   dao.OrderDAO
	cache cache.OrderCache
}

func NewOrderRepository(dao dao.OrderDAO, cache cache.OrderCache) OrderRepository {
	return &orderRepository{
		dao:   dao,
		cache: cache,
	}
}

// CreateTempOrder 创建临时订单
func (repo *orderRepository) CreateTempOrder(ctx context.Context, order *model.Order) error {
	if err := repo.dao.CreateTempOrder(ctx, order); err != nil {
		return toRepositoryErr(err)
	}
	return nil
}

func (repo *orderRepository) PayTempOrder(ctx context.Context, orderID int64) error {
	if err := repo.dao.PayTempOrder(ctx, orderID); err != nil {
		return toRepositoryErr(err)
	}
	return nil
}

func (repo *orderRepository) CancelTempOrder(ctx context.Context, orderID int64) error {
	if err := repo.dao.CancelTempOrder(ctx, orderID); err != nil {
		return toRepositoryErr(err)
	}
	return nil
}

func (repo *orderRepository) GetTempOrder(ctx context.Context, uid int64) (*model.Order, error) {
	if order, err := repo.dao.GetTempOrder(ctx, uid); err != nil {
		return nil, toRepositoryErr(err)
	} else {
		return order, nil
	}
}

func (repo *orderRepository) RecycleTempOrder(ctx context.Context, uid, tempOrderID int64) (bool, error) {
	if ok, err := repo.dao.RecycleTempOrder(ctx, uid, tempOrderID); err != nil {
		return false, ErrServerInternal
	} else {
		return ok, nil
	}
}

func (repo *orderRepository) GetOrder(ctx context.Context, uid int64) (*model.Order, error) {
	order, err := repo.dao.Get(ctx, uid)
	if err != nil {
		return nil, ErrRecordNotFound
	}
	return order, nil
}
