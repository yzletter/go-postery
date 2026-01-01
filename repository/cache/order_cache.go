package cache

import "github.com/redis/go-redis/v9"

type redisOrderCache struct {
	client redis.UniversalClient
}

func NewOrderCache(client redis.UniversalClient) OrderCache {
	return &redisOrderCache{client: client}
}
