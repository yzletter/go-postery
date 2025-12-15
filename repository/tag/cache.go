package repository

import "github.com/redis/go-redis/v9"

type TagCacheRepository struct {
	redisClient redis.Cmdable
}

func NewTagCacheRepository(redisClient redis.Cmdable) *TagCacheRepository {
	return &TagCacheRepository{redisClient: redisClient}
}
