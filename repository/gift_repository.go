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
