package cache

import "github.com/redis/go-redis/v9"

type redisGiftCache struct {
	client redis.UniversalClient
}

func NewGiftCache(client redis.UniversalClient) GiftCache {
	return &redisGiftCache{client: client}
}
