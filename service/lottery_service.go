package service

import (
	"github.com/yzletter/go-postery/repository"
	"github.com/yzletter/go-postery/service/ports"
)

type lotteryService struct {
	giftRepo  repository.GiftRepository
	orderRepo repository.OrderRepository
	mq        ports.LotteryMQ
}

func NewLotteryService(orderRepo repository.OrderRepository, giftRepo repository.GiftRepository, mq ports.LotteryMQ) LotteryService {
	return &lotteryService{
		orderRepo: orderRepo,
		giftRepo:  giftRepo,
		mq:        mq,
	}
}
