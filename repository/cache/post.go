package cache

import "github.com/go-redis/redis"

type RedisPostCache struct {
	client redis.Cmdable
}

func NewPostCache(client redis.Cmdable) *RedisPostCache {
	return &RedisPostCache{client: client}
}
