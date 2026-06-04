package cache

import (
	"context"
	_ "embed"
	"log/slog"
	"strconv"

	"github.com/redis/go-redis/v9"
	"github.com/yzletter/go-postery/microservice-backend/lottery/model"
)

const (
	lotteryGiftPrefix      = "lottery:gift:"
	stockRollbackPrefix    = "lottery:stock_rollback:"
	stockRollbackMarkerTTL = 0
)

//go:embed lua/decrease_inventory_script.lua
var luaDecrInventoryScript string

//go:embed lua/rollback_inventory_script.lua
var luaRollbackInventoryScript string

type redisGiftCache struct {
	client redis.UniversalClient
}

func NewGiftCache(client redis.UniversalClient) GiftCache {
	return &redisGiftCache{client: client}
}

func (cache *redisGiftCache) InitInventory(ctx context.Context, gifts []*model.Gift) {
	for _, gift := range gifts {
		if gift.Count <= 0 {
			slog.Error("Gift Count Invaild", "gift", gift)
			continue
		}

		// 初始化
		err := cache.client.Set(ctx, lotteryGiftPrefix+strconv.FormatInt(gift.ID, 10), gift.Count, 0).Err()
		if err != nil {
			slog.Error("Set Failed", "error", err)
		}
	}
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
	key := lotteryGiftPrefix + strconv.FormatInt(gid, 10)
	// 执行 Lua 脚本扣减库存
	if result, err := cache.client.Eval(ctx, luaDecrInventoryScript, []string{key}).Int(); err != nil {
		return err // Redis 错误
	} else if result != 1 {
		return ErrReduceInventory // 库存不足
	}

	return nil
}

func (cache *redisGiftCache) IncreaseInventory(ctx context.Context, gid int64) error {
	return cache.client.Incr(ctx, lotteryGiftPrefix+strconv.FormatInt(gid, 10)).Err()
}

func (cache *redisGiftCache) RollbackInventory(ctx context.Context, orderID, gid int64) error {
	result, err := cache.client.Eval(ctx, luaRollbackInventoryScript, []string{giftKey(gid), stockRollbackKey(orderID)}, stockRollbackMarkerTTL).Int()
	if err != nil {
		return err
	}
	if result != 1 && result != 2 {
		return ErrRollbackInventory
	}
	return nil
}

func stockRollbackKey(orderID int64) string {
	return stockRollbackPrefix + strconv.FormatInt(orderID, 10)
}
