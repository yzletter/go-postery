package cache

import (
	"context"
	"strconv"

	"github.com/redis/go-redis/v9"
)

const (
	lotteryOrderPrefix = "lottery:order:"
)

type redisOrderCache struct {
	client redis.UniversalClient
}

func NewOrderCache(client redis.UniversalClient) OrderCache {
	return &redisOrderCache{client: client}
}

// CreateTempOrder 创建临时订单
func (cache *redisOrderCache) CreateTempOrder(ctx context.Context, uid, gid int64) error {
	return cache.client.Set(ctx, lotteryOrderPrefix+strconv.FormatInt(uid, 10), strconv.FormatInt(gid, 10), 0).Err()
}

// DeleteTempOrder 删除临时订单
func (cache *redisOrderCache) DeleteTempOrder(ctx context.Context, uid int64) error {
	return cache.client.Del(ctx, lotteryOrderPrefix+strconv.FormatInt(uid, 10)).Err()
}

// GetTempOrderID 获取临时订单
func (cache *redisOrderCache) GetTempOrderID(ctx context.Context, uid int64) (int64, error) {
	return cache.client.Get(ctx, lotteryOrderPrefix+strconv.FormatInt(uid, 10)).Int64()
}
