package cache

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/yzletter/go-postery/microservice-backend/lottery/model"
)

const (
	lotteryTempOrderPrefix = "lottery:temp_order:"
	tempOrderTTL           = 15 * time.Minute
)

//go:embed lua/delete_temp_order_script.lua
var luaDeleteTempOrderScript string

//go:embed lua/recycle_temp_order_script.lua
var luaRecycleTempOrderScript string

type redisOrderCache struct {
	client redis.UniversalClient
}

func NewOrderCache(client redis.UniversalClient) OrderCache {
	return &redisOrderCache{client: client}
}

// CreateTempOrder 创建临时订单
func (cache *redisOrderCache) CreateTempOrder(ctx context.Context, order *model.TempOrder) error {
	body, err := json.Marshal(order)
	if err != nil {
		return err
	}
	if ok, err := cache.client.SetNX(ctx, tempOrderKey(order.UserID), body, tempOrderTTL).Result(); err != nil {
		return err
	} else if !ok {
		return ErrCreateTempOrder
	}
	return nil
}

// DeleteTempOrder 删除临时订单
func (cache *redisOrderCache) DeleteTempOrder(ctx context.Context, uid, tempOrderID int64) error {
	result, err := cache.client.Eval(ctx, luaDeleteTempOrderScript, []string{tempOrderKey(uid)}, strconv.FormatInt(tempOrderID, 10)).Int()
	if err != nil {
		return err
	}
	if result != 1 {
		return ErrTempOrderMissing
	}
	return nil
}

// RecycleTempOrder 回收临时订单并恢复库存，返回值表示消息是否可以 Ack。
func (cache *redisOrderCache) RecycleTempOrder(ctx context.Context, uid, tempOrderID int64) (bool, error) {
	// 获取 giftID
	tempOrder, err := cache.GetTempOrder(ctx, uid)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return true, nil
		}
		return false, err
	}
	if tempOrder.ID != tempOrderID {
		return true, nil
	}

	giftID := strconv.FormatInt(tempOrder.GiftID, 10)
	result, err := cache.client.Eval(ctx, luaRecycleTempOrderScript,
		[]string{tempOrderKey(uid), giftKey(tempOrder.GiftID)}, strconv.FormatInt(tempOrderID, 10), giftID).Int()
	if err != nil {
		return false, err
	}

	if result == 0 {
		return true, nil
	} else if result != 1 {
		return false, ErrRecycleTempOrder
	}
	return true, nil
}

// GetTempOrder 获取临时订单
func (cache *redisOrderCache) GetTempOrder(ctx context.Context, uid int64) (*model.TempOrder, error) {
	body, err := cache.client.Get(ctx, tempOrderKey(uid)).Bytes()
	if err != nil {
		return nil, err
	}

	var order model.TempOrder
	if err = json.Unmarshal(body, &order); err != nil {
		return nil, err
	}
	return &order, nil
}

func tempOrderKey(uid int64) string {
	return lotteryTempOrderPrefix + strconv.FormatInt(uid, 10)
}

func giftKey(gid int64) string {
	return lotteryGiftPrefix + strconv.FormatInt(gid, 10)
}
