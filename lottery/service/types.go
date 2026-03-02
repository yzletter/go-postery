package service

import (
	"context"

	"github.com/yzletter/go-postery/lottery/model"
)

type LotteryService interface {
	GetAllGifts(ctx context.Context) ([]*model.Gift, error)                      // 获取索引奖品
	Lottery(ctx context.Context, userID int64) (*model.Gift, error)              // 进行抽奖
	Pay(ctx context.Context, userID int64, giftID int64) error                   // 支付
	GiveUp(ctx context.Context, userID int64, giftID int64) error                // 放弃
	Result(ctx context.Context, userID int64) (*model.Order, *model.Gift, error) // 查询结果
	StartLotteryOrderConsumer(ctx context.Context)                               // 开启协程核查临时订单进行库存回流
	InitCacheInventory(ctx context.Context)                                      // 初始化库存
}
