package repository

import "github.com/go-redis/redis"

type CommentCacheRepository struct {
	redisClient redis.Cmdable
}

func NewCommentCacheRepository(redisClient redis.Cmdable) *CommentCacheRepository {
	return &CommentCacheRepository{
		redisClient: redisClient,
	}
}
