package repository

import (
	"context"

	"github.com/yzletter/go-postery/repository/cache"
)

type smsRepository struct {
	cache cache.SmsCache
}

func NewSmsRepository(cache cache.SmsCache) SmsRepository {
	return &smsRepository{cache: cache}
}

func (repo *smsRepository) CheckCode(ctx context.Context, phoneNumber string, code string) error {
	result, err := repo.cache.CheckCode(ctx, phoneNumber, code)
	if err != nil || result == -1 {
		return ErrServerInternal
	} else if result == 0 {
		return ErrResourceConflict
	}

	return nil
}
