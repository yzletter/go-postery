package repository

import (
	"context"

	"github.com/yzletter/go-postery/repository/cache"
)

type emailRepository struct {
	cache cache.EmailCache
}

func NewEmailRepository(cache cache.EmailCache) EmailRepository {
	return &emailRepository{
		cache: cache,
	}
}

func (repo *emailRepository) CheckCode(ctx context.Context, emailAddress string, code string) error {
	result, err := repo.cache.CheckCode(ctx, emailAddress, code)
	if err != nil || result == -1 {
		return ErrServerInternal
	} else if result == 0 {
		return ErrResourceConflict
	}

	return nil
}
