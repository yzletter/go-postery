package service

import (
	"context"

	lottery_grpc "github.com/yzletter/go-postery/api/proto/lottery/v1"
)

type LotteryService interface {
	GetAllGifts(context.Context, *lottery_grpc.EmptyRequest) (*lottery_grpc.Gifts, error)            // 获取索引奖品
	Lottery(context.Context, *lottery_grpc.UserID) (*lottery_grpc.Gift, error)                       // 进行抽奖
	Pay(context.Context, *lottery_grpc.LotteryCommonRequest) (*lottery_grpc.EmptyResponse, error)    // 支付
	GiveUp(context.Context, *lottery_grpc.LotteryCommonRequest) (*lottery_grpc.EmptyResponse, error) // 放弃
	Result(context.Context, *lottery_grpc.UserID) (*lottery_grpc.Order, error)                       // 查询结果
	StartLotteryOrderConsumer(ctx context.Context)                                                   // 开启协程核查临时订单进行库存回流
	InitCacheInventory(ctx context.Context)                                                          // 初始化库存
	lottery_grpc.UnsafeLotteryServiceServer
}
