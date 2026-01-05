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
	panic("todo")
}
