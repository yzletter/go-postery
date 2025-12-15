package repository

import (
	"github.com/yzletter/go-postery/repository/cache"
	"github.com/yzletter/go-postery/repository/dao"
)

type likeRepository struct {
	dao   dao.LikeDAO
	cache cache.LikeCache
}

func NewLikeRepository(likeDAO dao.LikeDAO, likeCache cache.LikeCache) LikeRepository {
	return &likeRepository{dao: likeDAO, cache: likeCache}
}
