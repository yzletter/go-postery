package cache

import "github.com/redis/go-redis/v9"

// redisLikeCache 用 Redis 实现 LikeCache
type redisLikeCache struct {
	client redis.Cmdable
}

// NewLikeCache 构造函数
func NewLikeCache(redisClient redis.Cmdable) LikeCache {
	return &redisLikeCache{client: redisClient}
}
