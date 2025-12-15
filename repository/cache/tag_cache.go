package cache

import "github.com/redis/go-redis/v9"

// RedisTagCache 用 Redis 实现 TagCache
type RedisTagCache struct {
	client redis.Cmdable
}

// NewTagCache 构造函数
func NewTagCache(redisClient redis.Cmdable) TagCache {
	return &RedisTagCache{client: redisClient}
}
