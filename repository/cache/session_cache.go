package cache

import "github.com/redis/go-redis/v9"

type redisSessionCache struct {
	client redis.UniversalClient
}

func NewSessionCache(client redis.UniversalClient) SessionCache {
	return &redisSessionCache{client: client}
}
