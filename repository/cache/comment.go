package cache

import "github.com/redis/go-redis/v9"

// RedisCommentCache 用 Redis 实现 CommentCache
type RedisCommentCache struct {
	client redis.Cmdable
}

// NewCommentCache 构造函数
func NewCommentCache(redisClient redis.Cmdable) CommentCache {
	return &RedisCommentCache{client: redisClient}
}
