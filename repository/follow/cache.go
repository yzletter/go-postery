package repository

import "github.com/go-redis/redis"

type FollowCacheRepository struct {
	redisClient redis.Cmdable
}

func NewFollowCacheRepository(redisClient redis.Cmdable) *FollowCacheRepository {
	return &FollowCacheRepository{redisClient: redisClient}
}
