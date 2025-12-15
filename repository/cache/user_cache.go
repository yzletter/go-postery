package cache

import "github.com/redis/go-redis/v9"

// redisUserCache 用 Redis 实现 UserCache
type redisUserCache struct {
	client redis.Cmdable
}

// NewUserCache 构造函数
func NewUserCache(client redis.Cmdable) UserCache {
	return &redisUserCache{client: client}
}
