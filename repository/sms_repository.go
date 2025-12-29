package repository

import "github.com/yzletter/go-postery/repository/cache"

type smsRepository struct {
	cache cache.SmsCache
}

func NewSmsRepository(cache cache.SmsCache) SmsRepository {
	return &smsRepository{cache: cache}
}
