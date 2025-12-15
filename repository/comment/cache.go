package repository

import "github.com/redis/go-redis/v9"

type CommentCacheRepository struct {
	redisClient redis.Cmdable
}

func NewCommentCacheRepository(redisClient redis.Cmdable) *CommentCacheRepository {
	return &CommentCacheRepository{
		redisClient: redisClient,
	}
}
