package cache

import (
	"context"
	_ "embed"
	"log/slog"
	"strconv"

	"github.com/redis/go-redis/v9"
	"github.com/yzletter/go-postery/backend/micro/lottery/model"
)

const (
	lotteryRecordPrefix = "lottery:record:"
	lotteryGiftPrefix   = "lottery:gift:"
	stockRollbackPrefix = "lottery:stock_rollback:"
)

//go:embed lua/decrease_inventory_script.lua
var luaDecrInventoryScript string

//go:embed lua/rollback_inventory_script.lua
var luaRollbackInventoryScript string

//go:embed lua/lottery_script.lua
var luaLotteryScript string

//go:embed lua/reuse_script.lua
var luaReuseScript string

type redisGiftCache struct {
	client redis.UniversalClient
}

func (cache *redisGiftCache) Lottery(ctx context.Context, requestID int64, userID int64, giftID int64) (int64, int64, error) {
	if requestID <= 0 || userID <= 0 || giftID <= 0 {
		return 0, 0, ErrInvalidParameter
	}

	res, err := cache.client.Eval(ctx, luaLotteryScript, []string{giftKey(giftID), recordKey(userID)}, requestID, userID, giftID).Slice()
	if err != nil {
		// 保留 Redis 错误, 不能当作库存不足
		return 0, 0, err
	}

	if len(res) != 3 {
		return 0, 0, ErrInvalidParameter
	}

	code, ok := res[0].(int64)
	if !ok {
		return 0, 0, ErrInvalidParameter
	}

	switch code {
	case 0:
		return 0, 0, ErrReduceInventory
	case 1, 2:
		// 新预扣和复用都成功, 解析实际预占记录
	default:
		return 0, 0, ErrInvalidParameter
	}

	return parseResult(res)
}

func (cache *redisGiftCache) Reuse(ctx context.Context, requestID int64, userID int64) (int64, int64, error) {
	if requestID <= 0 || userID <= 0 {
		return 0, 0, ErrInvalidParameter
	}

	res, err := cache.client.Eval(ctx, luaReuseScript, []string{recordKey(userID)}, requestID, userID).Slice()
	if err != nil {
		// 保留 Redis 错误
		return 0, 0, err
	}

	if len(res) != 3 {
		return 0, 0, ErrInvalidParameter
	}

	code, ok := res[0].(int64)
	if !ok {
		return 0, 0, ErrInvalidParameter
	}

	switch code {
	case 0:
		return 0, 0, ErrRecordNotFound
	case 1:
		// 复用记录成功
	default:
		return 0, 0, ErrInvalidParameter
	}

	return parseResult(res)
}

// ScanLotteryRecords 分页读取抽奖预占记录
func (cache *redisGiftCache) ScanLotteryRecords(ctx context.Context, cursor uint64, count int64) ([]*model.LotteryRecord, uint64, error) {
	keys, nextCursor, err := cache.client.Scan(ctx, cursor, lotteryRecordPrefix+"*", count).Result()
	if err != nil {
		return nil, cursor, err
	}

	var records []*model.LotteryRecord
	for _, key := range keys {
		data, err := cache.client.HGetAll(ctx, key).Result()
		if err != nil {
			slog.Error("get lottery record failed", "key", key, "error", err)
			continue
		}
		if len(data) == 0 {
			continue
		}

		requestID, requestErr := strconv.ParseInt(data["request_id"], 10, 64)
		userID, userErr := strconv.ParseInt(data["user_id"], 10, 64)
		giftID, giftErr := strconv.ParseInt(data["gift_id"], 10, 64)
		if requestErr != nil || userErr != nil || giftErr != nil ||
			requestID <= 0 || userID <= 0 || giftID <= 0 ||
			recordKey(userID) != key {
			slog.Warn("skip invalid lottery record", "key", key)
			continue
		}

		records = append(records, &model.LotteryRecord{
			RequestID: requestID,
			UserID:    userID,
			GiftID:    giftID,
		})
	}

	return records, nextCursor, nil
}

func parseResult(args []interface{}) (int64, int64, error) {
	// 断言成字符串
	requestIDText, requestOK := args[1].(string)
	giftIDText, giftOK := args[2].(string)

	// 非法数据
	if !requestOK || !giftOK {
		return 0, 0, ErrInvalidParameter
	}

	// 字符串转 int64
	actualRequestID, err := strconv.ParseInt(requestIDText, 10, 64)
	if err != nil || actualRequestID <= 0 {
		return 0, 0, ErrInvalidParameter
	}

	actualGiftID, err := strconv.ParseInt(giftIDText, 10, 64)
	if err != nil || actualGiftID <= 0 {
		return 0, 0, ErrInvalidParameter
	}
	return actualRequestID, actualGiftID, nil

}

func NewGiftCache(client redis.UniversalClient) GiftCache {
	return &redisGiftCache{client: client}
}

func (cache *redisGiftCache) InitInventory(ctx context.Context, gifts []*model.Gift) {
	for _, gift := range gifts {
		if gift.Count <= 0 {
			slog.Warn("skip invalid gift inventory", "gift_id", gift.ID, "count", gift.Count)
			continue
		}

		// 初始化
		err := cache.client.Set(ctx, lotteryGiftPrefix+strconv.FormatInt(gift.ID, 10), gift.Count, 0).Err()
		if err != nil {
			slog.Warn("set gift inventory cache failed", "gift_id", gift.ID, "error", err)
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

func (cache *redisGiftCache) RollbackInventory(ctx context.Context, orderID, gid int64) error {
	result, err := cache.client.Eval(ctx, luaRollbackInventoryScript, []string{giftKey(gid), stockRollbackKey(orderID)}).Int()
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

func recordKey(userID int64) string {
	return lotteryRecordPrefix + strconv.FormatInt(userID, 10)
}

func giftKey(gid int64) string {
	return lotteryGiftPrefix + strconv.FormatInt(gid, 10)
}
