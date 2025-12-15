package repository

import (
	"github.com/yzletter/go-postery/repository/cache"
	"github.com/yzletter/go-postery/repository/dao"
)

type followRepository struct {
	dao   dao.FollowDAO
	cache cache.FollowCache
}

func NewFollowRepository(followDAO dao.FollowDAO, followCache cache.FollowCache) FollowRepository {
	return &followRepository{dao: followDAO, cache: followCache}
}
