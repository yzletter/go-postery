package cache

import "github.com/redis/go-redis/v9"

// redisTagCache 用 Redis 实现 TagCache
type redisTagCache struct {
	client redis.Cmdable
}

// NewTagCache 构造函数
func NewTagCache(redisClient redis.Cmdable) TagCache {
	return &redisTagCache{client: redisClient}
}
