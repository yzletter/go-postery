package repository

import "github.com/go-redis/redis"

type LikeCacheRepository struct {
	redisClient redis.Cmdable
}

func NewLikeCacheRepository(redisClient redis.Cmdable) *LikeCacheRepository {
	return &LikeCacheRepository{redisClient: redisClient}
}
