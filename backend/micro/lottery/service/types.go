package service

import (
	"context"

	"github.com/yzletter/go-postery/backend/micro/lottery/dto"
	"github.com/yzletter/go-postery/backend/micro/lottery/model"
)

type LotteryService interface {
	// GetAllGifts 获取索引奖品
	//
	// Return:
	//	- []*model.Gift: 奖品列表
	//	- error: 可能返回的错误
	GetAllGifts(ctx context.Context) ([]*model.Gift, error)

	// Lottery 进行抽奖
	//
	// Parameter:
	//	- userID: 用户 ID
	//
	// Return:
	//	- *dto.LotteryResult: 抽奖结果
	//	- error: 可能返回的错误
	Lottery(ctx context.Context, userID int64) (*dto.LotteryResult, error)

	// Pay 支付
	//
	// Parameter:
	//	- userID: 用户 ID
	//	- tempOrderID: 临时订单 ID
	//	- giftID: 奖品 ID
	//
	// Return:
	//	- error: 可能返回的错误
	Pay(ctx context.Context, userID int64, tempOrderID int64, giftID int64) error

	// GiveUp 放弃
	//
	// Parameter:
	//	- userID: 用户 ID
	//	- tempOrderID: 临时订单 ID
	//	- giftID: 奖品 ID
	//
	// Return:
	//	- error: 可能返回的错误
	GiveUp(ctx context.Context, userID int64, tempOrderID int64, giftID int64) error

	// Result 查询结果
	//
	// Parameter:
	//	- userID: 用户 ID
	//
	// Return:
	//	- *model.Order: 订单
	//	- *model.Gift: 奖品
	//	- error: 可能返回的错误
	Result(ctx context.Context, userID int64) (*model.Order, *model.Gift, error)

	// StartLotteryOrderConsumer 开启协程核查临时订单进行库存回流
	//
	// Parameter:
	//	- ctx: 上下文
	StartLotteryOrderConsumer(ctx context.Context)

	// StartStockRollbackScanner 开启协程扫描失败库存回补
	//
	// Parameter:
	//	- ctx: 上下文
	StartStockRollbackScanner(ctx context.Context)

	// InitCacheInventory 初始化库存
	//
	// Parameter:
	//	- ctx: 上下文
	InitCacheInventory(ctx context.Context)
}
