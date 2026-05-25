package repository

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
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

func (repo *orderRepository) CreateTempOrder(ctx context.Context, uid, gid int64) error {
	if err := repo.cache.CreateTempOrder(ctx, uid, gid); err != nil {
		if errors.Is(err, cache.ErrCreateTempOrder) {
			return ErrResourceConflict
		}
		return ErrServerInternal
	}
	return nil
}

func (repo *orderRepository) DeleteTempOrder(ctx context.Context, uid int64) error {
	err := repo.cache.DeleteTempOrder(ctx, uid)
	if err != nil {
		return ErrServerInternal
	}

	return nil
}

func (repo *orderRepository) GetTempOrder(ctx context.Context, uid int64) (int64, error) {
	id, err := repo.cache.GetTempOrderID(ctx, uid)
	if err != nil {
		return 0, ErrRecordNotFound
	}
	return id, nil
}

func (repo *orderRepository) CheckTempOrder(ctx context.Context, uid int64) (bool, error) {
	_, err := repo.cache.GetTempOrderID(ctx, uid)
	if err == nil {
		return true, nil
	} else if errors.Is(err, redis.Nil) {
		return false, nil
	}
	return false, ErrServerInternal
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
