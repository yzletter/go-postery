package cache

import "github.com/redis/go-redis/v9"

type redisMessageCache struct {
	client redis.UniversalClient
}

func NewMessageCache(client redis.UniversalClient) MessageCache {
	return &redisMessageCache{client: client}
}
