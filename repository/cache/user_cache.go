package cache

import (
	"context"
	"strconv"

	"github.com/redis/go-redis/v9"
	"github.com/yzletter/go-postery/model"
)

// redisUserCache 用 Redis 实现 UserCache
type redisUserCache struct {
	client redis.UniversalClient
}

// NewUserCache 构造函数
func NewUserCache(client redis.UniversalClient) UserCache {
	return &redisUserCache{client: client}
}
func (cache *redisUserCache) Top(ctx context.Context) ([]int64, []float64, error) {
	pairs, err := cache.client.ZRevRangeWithScores(ctx, model.KeyUserScore, 0, 5).Result()
	if err != nil {
		return nil, nil, err
	}
	var ids []int64
	var scores []float64
	for _, pair := range pairs {
		id, err := strconv.ParseInt(pair.Member.(string), 10, 64)
		if err != nil {
			id = 0
		}
		score := pair.Score
		ids = append(ids, id)
		scores = append(scores, score)
	}

	return ids, scores, nil
}

func (cache *redisUserCache) ChangeScore(ctx context.Context, pid int64, delta int) error {
	_, err := cache.client.ZIncrBy(ctx, model.KeyUserScore, float64(delta), strconv.FormatInt(pid, 10)).Result()
	return err
}
