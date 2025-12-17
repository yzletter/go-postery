package cache

import "github.com/redis/go-redis/v9"

// redisCommentCache 用 Redis 实现 CommentCache
type redisCommentCache struct {
	client redis.UniversalClient
}

// NewCommentCache 构造函数
func NewCommentCache(redisClient redis.UniversalClient) CommentCache {
	return &redisCommentCache{client: redisClient}
}
