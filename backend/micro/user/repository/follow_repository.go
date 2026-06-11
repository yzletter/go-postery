package repository

import (
	"context"

	"github.com/yzletter/go-postery/backend/micro/user/model"
	"github.com/yzletter/go-postery/backend/micro/user/repository/dao"
)

type followRepository struct {
	dao dao.FollowDAO
}

func NewFollowRepository(followDAO dao.FollowDAO) FollowRepository {
	return &followRepository{dao: followDAO}
}

func (repo *followRepository) Create(ctx context.Context, follow *model.Follow) error {
	err := repo.dao.Create(ctx, follow)
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

func (repo *followRepository) Exists(ctx context.Context, ferID, feeID int64) (model.FollowType, error) {
	res, err := repo.dao.Exists(ctx, ferID, feeID)
	if err != nil {
		return 0, toRepositoryErr(err)
	}

	// todo 写 Cache

	return res, nil
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
