package repository

import "github.com/redis/go-redis/v9"

type FollowCacheRepository struct {
	redisClient redis.Cmdable
}

func NewFollowCacheRepository(redisClient redis.Cmdable) *FollowCacheRepository {
	return &FollowCacheRepository{redisClient: redisClient}
}
