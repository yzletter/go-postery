package repository

import (
	"context"

	"github.com/yzletter/go-postery/model"
	"github.com/yzletter/go-postery/repository/cache"
	"github.com/yzletter/go-postery/repository/dao"
)

type giftRepository struct {
	dao   dao.GiftDAO
	cache cache.GiftCache
}

func NewGiftRepository(dao dao.GiftDAO, cache cache.GiftCache) GiftRepository {
	return &giftRepository{
		dao:   dao,
		cache: cache,
	}
}

func (repo *giftRepository) GetAllGifts(ctx context.Context) ([]*model.Gift, error) {
	gifts, err := repo.dao.GetAll(ctx)
	if err != nil {
		return nil, toRepositoryErr(err)
	}
	return gifts, nil
}

func (repo *giftRepository) GetCacheInventory(ctx context.Context) ([]*model.Gift, error) {
	gifts, err := repo.cache.GetAllInventory(ctx)
	if err != nil {
		return nil, ErrRecordNotFound
	}
	return gifts, nil
}

func (repo *giftRepository) GetByID(ctx context.Context, gid int64) (*model.Gift, error) {
	gift, err := repo.dao.GetByID(ctx, gid)
	if err != nil {
		return nil, toRepositoryErr(err)
	}
	return gift, nil
}

func (repo *giftRepository) ReduceCacheInventory(ctx context.Context, gid int64) error {
	err := repo.cache.ReduceInventory(ctx, gid)
	if err != nil {
		return ErrResourceConflict
	}
	return nil
}

func (repo *giftRepository) IncreaseCacheInventory(ctx context.Context, gid int64) error {
	err := repo.cache.IncreaseInventory(ctx, gid)
	if err != nil {
		return ErrResourceConflict
	}
	return nil
}
