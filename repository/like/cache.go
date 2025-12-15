package repository

import "github.com/redis/go-redis/v9"

type LikeCacheRepository struct {
	redisClient redis.Cmdable
}

func NewLikeCacheRepository(redisClient redis.Cmdable) *LikeCacheRepository {
	return &LikeCacheRepository{redisClient: redisClient}
}
