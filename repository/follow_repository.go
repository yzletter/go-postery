package repository

import (
	"context"

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

func (repo *followRepository) Create(ctx context.Context, ferID, feeID int64) error {
	err := repo.dao.Create(ctx, ferID, feeID)
	if err != nil {
		return toRepositoryErr(err)
	}

	return nil
}

func (repo *followRepository) Delete(ctx context.Context, ferID, feeID int64) error {
	err := repo.dao.Delete(ctx, ferID, feeID)
	if err != nil {
		return toRepositoryErr(err)
	}

	return nil
}

func (repo *followRepository) Exists(ctx context.Context, ferID, feeID int64) (int, error) {
	ok, err := repo.dao.Exists(ctx, ferID, feeID)
	if err != nil {
		return 0, toRepositoryErr(err)
	}

	// todo 写 Cache

	return ok, nil
}

func (repo *followRepository) GetFollowers(ctx context.Context, id int64, pageNo, pageSize int) (int64, []int64, error) {
	total, ids, err := repo.dao.GetFollowers(ctx, id, pageNo, pageSize)
	if err != nil {
		return 0, nil, toRepositoryErr(err)
	}

	// todo 写 Cache

	return total, ids, nil
}

func (repo *followRepository) GetFollowees(ctx context.Context, id int64, pageNo, pageSize int) (int64, []int64, error) {
	total, ids, err := repo.dao.GetFollowees(ctx, id, pageNo, pageSize)
	if err != nil {
		return 0, nil, toRepositoryErr(err)
	}

	// todo 写 Cache

	return total, ids, nil
}
