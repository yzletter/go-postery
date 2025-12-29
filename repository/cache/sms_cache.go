package cache

import "github.com/redis/go-redis/v9"

type redisSmsCache struct {
	client redis.UniversalClient
}

func NewSmsCache(client redis.UniversalClient) SmsCache {
	return redisSmsCache{client: client}
}
