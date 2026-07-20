package cache

import (
	"context"
	"strconv"

	"github.com/redis/go-redis/v9"
	"github.com/yzletter/go-postery/backend/micro/rank/domain"
)

const (
	rankUserZSetKey = "rank:user"
	rankPostZSetKey = "rank:post"
)

type redisRankCache struct {
	client redis.UniversalClient
}

func NewRankCache(client redis.UniversalClient) RankCache {
	return &redisRankCache{
		client: client,
	}
}

// DeleteScore 删除分数
func (cache *redisRankCache) DeleteScore(ctx context.Context, biz int, bizID int64) error {
	var key string
	if biz == domain.BizUser {
		key = rankUserZSetKey
	} else {
		key = rankPostZSetKey
	}

	return cache.client.ZRem(ctx, key, strconv.FormatInt(bizID, 10)).Err()
}

// UpdateScore 更新分数
func (cache *redisRankCache) UpdateScore(ctx context.Context, biz int, bizID int64, score int64) error {
	var key string
	if biz == domain.BizUser {
		key = rankUserZSetKey
	} else {
		key = rankPostZSetKey
	}

	if err := cache.client.ZAdd(ctx, key, redis.Z{Score: float64(score), Member: strconv.FormatInt(bizID, 10)}).Err(); err != nil {
		return err
	}
	return nil
}

// TopK 获取排名
func (cache *redisRankCache) TopK(ctx context.Context, biz int, k int) ([]int64, []int64, error) {
	var key string
	if biz == domain.BizUser {
		key = rankUserZSetKey
	} else {
		key = rankPostZSetKey
	}

	// 获取键值对
	pairs, err := cache.client.ZRevRangeWithScores(ctx, key, 0, int64(k-1)).Result()
	if err != nil {
		return nil, nil, err
	}

	// 转化
	ids := make([]int64, 0, len(pairs))
	scores := make([]int64, 0, len(pairs))
	for _, pair := range pairs {
		// ID
		id, err := strconv.ParseInt(pair.Member.(string), 10, 64)
		if err != nil {
			continue
		}

		ids = append(ids, id)
		scores = append(scores, int64(pair.Score))
	}

	return ids, scores, nil
}
