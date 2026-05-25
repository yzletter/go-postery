package cache

import (
	"context"
	_ "embed"
	"encoding/json"
	"strconv"

	"github.com/redis/go-redis/v9"
	"github.com/yzletter/go-postery/microservice-backend/lottery/model"
)

const (
	lotteryTempOrderPrefix = "lottery:temp_order:"
	tempOrderTTL           = 0
)

//go:embed lua/delete_temp_order_script.lua
var luaDeleteTempOrderScript string

type redisOrderCache struct {
	client redis.UniversalClient
}

func NewOrderCache(client redis.UniversalClient) OrderCache {
	return &redisOrderCache{client: client}
}

// CreateTempOrder 创建临时订单
func (cache *redisOrderCache) CreateTempOrder(ctx context.Context, order *model.TempOrder) error {
	value, err := json.Marshal(order)
	if err != nil {
		return err
	}

	if ok, err := cache.client.SetNX(ctx, tempOrderKey(order.UserID), value, tempOrderTTL).Result(); err != nil {
		return err
	} else if !ok {
		return ErrCreateTempOrder
	}
	return nil
}

// DeleteTempOrder 删除临时订单
func (cache *redisOrderCache) DeleteTempOrder(ctx context.Context, uid, tempOrderID int64) error {
	result, err := cache.client.Eval(
		ctx,
		luaDeleteTempOrderScript,
		[]string{tempOrderKey(uid)},
		strconv.FormatInt(tempOrderID, 10),
	).Int()
	if err != nil {
		return err
	}
	if result != 1 {
		return ErrTempOrderMissing
	}
	return nil
}

// GetTempOrder 获取临时订单
func (cache *redisOrderCache) GetTempOrder(ctx context.Context, uid int64) (*model.TempOrder, error) {
	value, err := cache.client.Get(ctx, tempOrderKey(uid)).Bytes()
	if err != nil {
		return nil, err
	}

	var order model.TempOrder
	if err = json.Unmarshal(value, &order); err != nil {
		return nil, err
	}
	return &order, nil
}

func tempOrderKey(uid int64) string {
	return lotteryTempOrderPrefix + strconv.FormatInt(uid, 10)
}
