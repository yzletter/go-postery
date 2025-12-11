package repository

import "github.com/go-redis/redis"

type PostCacheRepository struct {
	redisClient redis.Cmdable
}

func NewPostCacheRepository(redisClient redis.Cmdable) *PostCacheRepository {
	return &PostCacheRepository{
		redisClient: redisClient,
	}
}
