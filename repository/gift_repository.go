package repository

import (
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
