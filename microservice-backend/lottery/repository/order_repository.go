package repository

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
	"github.com/yzletter/go-postery/microservice-backend/lottery/dto"
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

func (repo *orderRepository) CreateTempOrder(ctx context.Context, order *dto.Order) error {
	if err := repo.cache.CreateTempOrder(ctx, order); err != nil {
		if errors.Is(err, cache.ErrCreateTempOrder) {
			return ErrResourceConflict
		}
		return ErrServerInternal
	}
	return nil
}

func (repo *orderRepository) DeleteTempOrder(ctx context.Context, uid, tempOrderID int64) error {
	if err := repo.cache.DeleteTempOrder(ctx, uid, tempOrderID); err != nil {
		if errors.Is(err, cache.ErrTempOrderMissing) {
			return ErrRecordNotFound
		}
		return ErrServerInternal
	}
	return nil
}

func (repo *orderRepository) RecycleTempOrder(ctx context.Context, uid, tempOrderID int64) (bool, error) {
	if ok, err := repo.cache.RecycleTempOrder(ctx, uid, tempOrderID); err != nil {
		return false, ErrServerInternal
	} else {
		return ok, nil
	}
}

func (repo *orderRepository) GetTempOrder(ctx context.Context, uid int64) (*dto.Order, error) {
	order, err := repo.cache.GetTempOrder(ctx, uid)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrRecordNotFound
		}
		return nil, ErrServerInternal
	}
	return order, nil
}

func (repo *orderRepository) CreateOrder(ctx context.Context, order *model.Order) error {
	err := repo.dao.Create(ctx, order)
	if err != nil {
		return toRepositoryErr(err)
	}
	return nil
}

func (repo *orderRepository) GetOrder(ctx context.Context, uid int64) (*model.Order, error) {
	order, err := repo.dao.Get(ctx, uid)
	if err != nil {
		return nil, ErrRecordNotFound
	}
	return order, nil
}
