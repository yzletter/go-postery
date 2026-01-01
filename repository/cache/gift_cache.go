package cache

import (
	"context"
	"strconv"

	"github.com/redis/go-redis/v9"
	"github.com/yzletter/go-postery/model"
)

const (
	lotteryGiftPrefix = "lottery:gift:"
)

type redisGiftCache struct {
	client redis.UniversalClient
}

func NewGiftCache(client redis.UniversalClient) GiftCache {
	return &redisGiftCache{client: client}
}

func (cache *redisGiftCache) InitInventory(ctx context.Context) (int, error) {
	//TODO implement me
	panic("implement me")
}

// GetAllInventory 获取缓存中所有奖品的库存量
func (cache *redisGiftCache) GetAllInventory(ctx context.Context) ([]*model.Gift, error) {
	// 获取所有 Key
	keys, err := cache.client.Keys(ctx, lotteryGiftPrefix+"*").Result()
	if err != nil {
		return nil, err
	}

	var gifts []*model.Gift
	for _, key := range keys {
		count, err := cache.client.Get(ctx, key).Int()
		if err != nil {
			continue
		}

		// 从 lottery:gift: 中获取 gid
		gid, err := strconv.ParseInt(key[len(lotteryGiftPrefix):], 10, 64)
		gift := &model.Gift{
			ID:    gid,
			Count: count,
		}

		gifts = append(gifts, gift)
	}

	return gifts, nil
}

func (cache *redisGiftCache) ReduceInventory(ctx context.Context, gid int64) error {
	count, err := cache.client.Decr(ctx, lotteryGiftPrefix+strconv.FormatInt(gid, 10)).Result()
	if err != nil {
		return err
	} else if count < 0 {
		return ErrReduceInventory
	}
	return nil
}

func (cache *redisGiftCache) IncreaseInventory(ctx context.Context, gid int64) error {
	return cache.client.Incr(ctx, lotteryGiftPrefix+strconv.FormatInt(gid, 10)).Err()
}
