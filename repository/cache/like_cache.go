package cache

import "github.com/redis/go-redis/v9"

// redisLikeCache 用 Redis 实现 LikeCache
type redisLikeCache struct {
	client redis.UniversalClient
}

// NewLikeCache 构造函数
func NewLikeCache(redisClient redis.UniversalClient) LikeCache {
	return &redisLikeCache{client: redisClient}
}
