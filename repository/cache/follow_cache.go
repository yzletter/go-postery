package cache

import "github.com/redis/go-redis/v9"

// redisFollowCache 用 Redis 实现 FollowCache
type redisFollowCache struct {
	client redis.Cmdable
}

// NewFollowCache 构造函数
func NewFollowCache(redisClient redis.Cmdable) FollowCache {
	return &redisFollowCache{client: redisClient}
}
