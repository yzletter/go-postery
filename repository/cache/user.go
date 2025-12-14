package cache

import "github.com/go-redis/redis"

// RedisUserCache 用 Redis 实现 UserCache
type RedisUserCache struct {
	client redis.Cmdable
}

// NewUserCache 构造函数
func NewUserCache(client redis.Cmdable) *RedisUserCache {
	return &RedisUserCache{client: client}
}
