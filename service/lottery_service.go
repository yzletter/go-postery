package service

import (
	"context"
	"errors"
	"math/rand"

	giftdto "github.com/yzletter/go-postery/dto/gift"
	orderdto "github.com/yzletter/go-postery/dto/order"
	"github.com/yzletter/go-postery/errno"
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

func (svc *lotteryService) GetAllGifts(ctx context.Context) ([]giftdto.DTO, error) {
	var empty []giftdto.DTO
	gifts, err := svc.giftRepo.GetAllGifts(ctx)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return empty, errno.ErrGiftNotFound
		}
		return empty, errno.ErrServerInternal
	}

	for _, gift := range gifts {
		giftDTO := giftdto.ToDTO(gift)
		empty = append(empty, giftDTO)
	}

	return empty, nil
}

func (svc *lotteryService) Lottery(ctx context.Context) (giftdto.DTO, error) {
	//TODO implement me
	panic("implement me")
}

func (svc *lotteryService) GiveUp(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (svc *lotteryService) Pay(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (svc *lotteryService) Result(ctx context.Context) ([]orderdto.DTO, error) {
	//TODO implement me
	panic("implement me")
}

// 抽奖算法
func lottery(probs []float64) int {
	if len(probs) == 0 {
		return -1
	}
	sum := 0.0

	acc := make([]float64, 0, len(probs))
	for _, prob := range probs {
		sum += prob
		acc = append(acc, sum)
	}

	// 获取 [0, sum) 的随机数
	x := rand.Float64() * sum

	// 大于等于 x 的第一个数的位置
	l, r := 0, len(probs)-1
	for l < r {
		mid := (l + r) / 2
		if acc[mid] < x {
			l = mid + 1
		} else {
			r = mid
		}
	}

	return l
}
