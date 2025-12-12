package repository

import "github.com/go-redis/redis"

type TagCacheRepository struct {
	redisClient redis.Cmdable
}

func NewTagCacheRepository(redisClient redis.Cmdable) *TagCacheRepository {
	return &TagCacheRepository{redisClient: redisClient}
}
