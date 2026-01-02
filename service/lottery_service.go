package service

import (
	"context"

	infraRocketMQ "github.com/yzletter/go-postery/infra/rocketmq"
	"github.com/yzletter/go-postery/repository"
)

type lotteryService struct {
	giftRepo  repository.GiftRepository
	orderRepo repository.OrderRepository
	mq        *infraRocketMQ.RocketMQ
}

func NewLotteryService(orderRepo repository.OrderRepository, giftRepo repository.GiftRepository, mq *infraRocketMQ.RocketMQ) LotteryService {
	return &lotteryService{
		orderRepo: orderRepo,
		giftRepo:  giftRepo,
		mq:        mq,
	}
}

func (svc *lotteryService) GetAllGifts(ctx context.Context) {

}
