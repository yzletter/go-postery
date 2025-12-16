package repository

import (
	"context"

	"github.com/yzletter/go-postery/model"
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

func (repo *likeRepository) Like(ctx context.Context, like *model.Like) error {
	err := repo.dao.Create(ctx, like)
	if err != nil {
		return toRepositoryErr(err)
	}
	return nil
}

func (repo *likeRepository) UnLike(ctx context.Context, uid, pid int64) error {
	err := repo.dao.Delete(ctx, uid, pid)
	if err != nil {
		return toRepositoryErr(err)
	}
	return nil
}

func (repo *likeRepository) HasLiked(ctx context.Context, uid, pid int64) (bool, error) {
	// todo 查 Cache
	ok, err := repo.dao.Exists(ctx, uid, pid)
	if err != nil {
		return false, toRepositoryErr(err)
	}

	return ok, nil
}
