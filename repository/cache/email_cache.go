package cache

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type redisEmailCache struct {
	client redis.UniversalClient
}

func NewEmailCache(client redis.UniversalClient) EmailCache {
	return &redisEmailCache{
		client: client,
	}
}

func (cache *redisEmailCache) CheckCode(ctx context.Context, emailAddress string, code string) (int, error) {
	panic("todo")
}
