package repository

import "github.com/go-redis/redis"

type UserCacheRepository struct {
	redisClient redis.Cmdable
}

func NewUserCacheRepository(redisClient redis.Cmdable) *UserCacheRepository {
	return &UserCacheRepository{
		redisClient: redisClient,
	}
}
