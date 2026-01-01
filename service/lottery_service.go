package service

import "github.com/yzletter/go-postery/repository"

type lotteryService struct {
	giftRepo  repository.GiftRepository
	orderRepo repository.OrderRepository
}

func NewLotteryService(orderRepo repository.OrderRepository, giftRepo repository.GiftRepository) LotteryService {
	return &lotteryService{
		orderRepo: orderRepo,
		giftRepo:  giftRepo,
	}
}
