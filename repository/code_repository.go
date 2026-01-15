package repository

import (
	"context"

	"github.com/yzletter/go-postery/model"
	"github.com/yzletter/go-postery/repository/cache"
)

type codeRepository struct {
	cache cache.CodeCache
}

func NewCodeRepository(cache cache.CodeCache) CodeRepository {
	return &codeRepository{
		cache: cache,
	}
}

func (repo *codeRepository) Allow(ctx context.Context, biz model.CodeBiz, field string, code string) error {
	result, err := repo.cache.Allow(ctx, biz, field, code)
	if err != nil || result == -1 {
		return ErrServerInternal
	} else if result == 0 {
		return ErrResourceConflict
	}
	return nil
}

func (repo *codeRepository) CheckCode(ctx context.Context, biz model.CodeBiz, field string, code string) (bool, error) {
	ok, err := repo.cache.CheckCode(ctx, biz, field, code)
	if err != nil {
		return false, toRepositoryErr(err)
	}

	return ok, nil
}
