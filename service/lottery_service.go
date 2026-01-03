package service

import (
	"context"
	"errors"
	"log/slog"
	"math/rand/v2"
	"time"

	rmq_client "github.com/apache/rocketmq-clients/golang/v5"
	"github.com/bytedance/sonic"
	"github.com/yzletter/go-postery/conf"
	giftdto "github.com/yzletter/go-postery/dto/gift"
	orderdto "github.com/yzletter/go-postery/dto/order"
	"github.com/yzletter/go-postery/errno"
	infraRocketMQ "github.com/yzletter/go-postery/infra/rocketmq"
	"github.com/yzletter/go-postery/model"
	"github.com/yzletter/go-postery/repository"
	"github.com/yzletter/go-postery/service/ports"
)

type lotteryService struct {
	giftRepo  repository.GiftRepository
	orderRepo repository.OrderRepository
	mq        *infraRocketMQ.RocketMQ
	idGen     ports.IDGenerator
}

func NewLotteryService(orderRepo repository.OrderRepository, giftRepo repository.GiftRepository, mq *infraRocketMQ.RocketMQ, idGen ports.IDGenerator) LotteryService {
	return &lotteryService{
		orderRepo: orderRepo,
		giftRepo:  giftRepo,
		mq:        mq,
		idGen:     idGen,
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

func (svc *lotteryService) Lottery(ctx context.Context, uid int64) (giftdto.DTO, error) {
	for try := 1; try <= 10; try++ {
		// 获取缓存中的库存
		gifts, err := svc.giftRepo.GetCacheInventory(ctx)
		if err != nil {
			continue
		}

		// 所有奖品已抽完
		if len(gifts) == 0 {
			empty := &model.Gift{
				ID:   0,
				Name: "奖品已抽完",
			}
			return giftdto.ToDTO(empty), nil
		}

		ids := make([]int64, len(gifts))
		probs := make([]float64, len(gifts))
		for i, gift := range gifts {
			ids[i] = gift.ID
			probs[i] = float64(gift.Count)
		}

		// 进行抽奖
		idx := lottery(probs)
		if idx == -1 {
			continue
		}

		gid := ids[idx]

		// 扣减缓存库存
		err = svc.giftRepo.ReduceCacheInventory(ctx, gid)
		if err != nil {
			// 扣减失败
			continue
		}

		// 获取商品详情
		gift, err := svc.giftRepo.GetByID(ctx, gid)
		if err != nil {
			// 获取不到详情
			_ = svc.giftRepo.IncreaseCacheInventory(ctx, gid)
			continue
		}

		// 创建临时订单
		err = svc.orderRepo.CreateTempOrder(ctx, uid, gid)
		if err != nil {
			_ = svc.giftRepo.IncreaseCacheInventory(ctx, gid)
			continue
		}

		// 发送延迟消息
		err = svc.produce(ctx, &model.Order{UserID: uid, GiftID: gid}, conf.RocketLotteryPayDelay)
		if err != nil {
			_ = svc.giftRepo.IncreaseCacheInventory(ctx, gid)
			_ = svc.orderRepo.DeleteTempOrder(ctx, uid)
			continue
		}

		// 返回数据
		return giftdto.ToDTO(gift), nil
	}

	empty := &model.Gift{
		ID:   1,
		Name: "谢谢参与",
	}
	return giftdto.ToDTO(empty), errno.ErrLotteryNoting
}

func (svc *lotteryService) produce(ctx context.Context, order *model.Order, delay int) error {
	// 序列化 Order
	body, err := sonic.Marshal(order)
	if err != nil {
		return errno.ErrServerInternal

	}

	// 构造 Message
	message := &rmq_client.Message{
		Topic: conf.RocketLotteryTopic,
		Body:  body,
	}
	message.SetDelayTimestamp(time.Now().Add(time.Duration(delay)))

	// 发送消息
	_, err = svc.mq.RocketProducer.Send(ctx, message)
	if err != nil {
		return errno.ErrServerInternal
	}
	return nil
}

func (svc *lotteryService) GiveUp(ctx context.Context, uid, gid int64) error {
	_ = svc.orderRepo.DeleteTempOrder(ctx, uid)
	_ = svc.giftRepo.IncreaseCacheInventory(ctx, gid)
	return nil
}

func (svc *lotteryService) Pay(ctx context.Context, uid, gid int64) error {
	// 获取临时订单
	tempID, err := svc.orderRepo.GetTempOrder(ctx, uid)
	if err != nil || tempID != gid {
		_ = svc.giftRepo.IncreaseCacheInventory(ctx, gid)
		return errno.ErrNotLottery
	}

	// 正式订单落库
	order := &model.Order{
		ID:     svc.idGen.NextID(),
		UserID: uid,
		GiftID: gid,
		Count:  1,
	}

	err = svc.orderRepo.CreateOrder(ctx, order)
	if err != nil {
		_ = svc.giftRepo.IncreaseCacheInventory(ctx, gid)
		if errors.Is(err, repository.ErrUniqueKey) {
			slog.Error("Create Order Failed", "error", err)
			return errno.ErrServerInternal
		}
		return errno.ErrServerInternal
	}

	// 删除临时订单
	_ = svc.orderRepo.DeleteTempOrder(ctx, uid)
	return nil
}

func (svc *lotteryService) Result(ctx context.Context, uid int64) (orderdto.DTO, error) {
	var empty orderdto.DTO
	order, err := svc.orderRepo.GetOrder(ctx, uid)
	if err != nil {
		return empty, errno.ErrOrderNotFound
	}

	return orderdto.ToDTO(order), nil
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
