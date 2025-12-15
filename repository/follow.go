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

func (repo *followRepository) Follow(ctx context.Context, ferID, feeID int64) error {
	err := repo.dao.Follow(ctx, ferID, feeID)
	if err != nil {
		return err
	}

	return nil
}

func (repo *followRepository) UnFollow(ctx context.Context, ferID, feeID int64) error {
	err := repo.dao.UnFollow(ctx, ferID, feeID)
	if err != nil {
		return err
	}

	return nil
}

func (repo *followRepository) IfFollow(ctx context.Context, ferID, feeID int64) (int, error) {
	ok, err := repo.dao.IfFollow(ctx, ferID, feeID)
	if err != nil {
		return 0, err
	}

	// todo 写 Cache

	return ok, nil
}

func (repo *followRepository) GetFollowers(ctx context.Context, id int64, pageNo, pageSize int) (int64, []int64, error) {
	total, ids, err := repo.dao.GetFollowers(ctx, id, pageNo, pageSize)
	if err != nil {
		return 0, nil, err
	}

	// todo 写 Cache

	return total, ids, nil
}

func (repo *followRepository) GetFollowees(ctx context.Context, id int64, pageNo, pageSize int) (int64, []int64, error) {
	total, ids, err := repo.dao.GetFollowees(ctx, id, pageNo, pageSize)
	if err != nil {
		return 0, nil, err
	}

	// todo 写 Cache

	return total, ids, nil
}
