package service

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"math/rand"
	"time"

	rmq_client "github.com/apache/rocketmq-clients/golang/v5"
	"github.com/bytedance/sonic"
	"github.com/yzletter/go-postery/microservice-backend/lottery/conf"
	"github.com/yzletter/go-postery/microservice-backend/lottery/errs"
	infraRocketMQ "github.com/yzletter/go-postery/microservice-backend/lottery/infra/rocketmq"
	model2 "github.com/yzletter/go-postery/microservice-backend/lottery/model"
	repository2 "github.com/yzletter/go-postery/microservice-backend/lottery/repository"
	"github.com/yzletter/go-postery/microservice-backend/lottery/service/ports"
)

type lotteryService struct {
	giftRepo  repository2.GiftRepository
	orderRepo repository2.OrderRepository
	mq        *infraRocketMQ.RocketMQ
	idGen     ports.IDGenerator
}

func NewLotteryService(orderRepo repository2.OrderRepository, giftRepo repository2.GiftRepository, mq *infraRocketMQ.RocketMQ, idGen ports.IDGenerator) LotteryService {
	return &lotteryService{
		orderRepo: orderRepo,
		giftRepo:  giftRepo,
		idGen:     idGen,
		mq:        mq,
	}
}

func (svc *lotteryService) GetAllGifts(ctx context.Context) ([]*model2.Gift, error) {
	gifts, err := svc.giftRepo.GetAllGifts(ctx)
	if err != nil {
		if errors.Is(err, repository2.ErrRecordNotFound) {
			slog.Error("Gift Not Found", "error", err)
			return nil, errs.ErrNotFound
		}
		slog.Error("Server Internal Error", "error", err)
		return nil, errs.ErrInternal
	}

	return gifts, nil
}

func (svc *lotteryService) Lottery(ctx context.Context, userID int64) (*model2.Gift, error) {
	for try := 1; try <= 10; try++ {
		// 获取缓存中的库存
		gifts, err := svc.giftRepo.GetCacheInventory(ctx)
		if err != nil {
			continue
		}

		ids := make([]int64, 0, len(gifts))
		probs := make([]float64, 0, len(gifts))

		// 只保留 > 0 的部分
		for _, gift := range gifts {
			if gift.Count > 0 {
				ids = append(ids, gift.ID)
				probs = append(probs, float64(gift.Count))
			}
		}

		// 所有奖品已抽完
		if len(probs) == 0 {
			empty := &model2.Gift{
				ID:   0,
				Name: "奖品已抽完",
			}
			return empty, nil
		}

		// 进行抽奖
		idx := lottery(probs)
		if idx == -1 {
			continue
		}

		gid := ids[idx]

		// 扣减缓存库存
		if err := svc.giftRepo.ReduceCacheInventory(ctx, gid); err != nil {
			// 扣减失败
			continue
		}

		// 获取商品详情
		gift, err := svc.giftRepo.GetByID(ctx, gid)
		if err != nil {
			// 获取不到详情
			_ = svc.giftRepo.IncreaseCacheInventory(ctx, gid)
			continue
		} else if gift.Name == "谢谢参与" {
			// 不用创建临时订单
			return gift, nil
		}

		// 创建临时订单
		if err := svc.orderRepo.CreateTempOrder(ctx, userID, gid); err != nil {
			_ = svc.giftRepo.IncreaseCacheInventory(ctx, gid)
			continue
		}

		// 发送延迟消息
		err = svc.produce(ctx, &model2.Order{UserID: userID, GiftID: gid}, conf.RocketLotteryPayDelay)
		if err != nil {
			_ = svc.giftRepo.IncreaseCacheInventory(ctx, gid)
			_ = svc.orderRepo.DeleteTempOrder(ctx, userID)
			continue
		}

		// 返回数据
		return gift, nil
	}

	empty := &model2.Gift{
		ID:   1,
		Name: "谢谢参与",
	}
	slog.Error("Lottery Nothing")
	return empty, nil
}

func (svc *lotteryService) Pay(ctx context.Context, userID int64, giftID int64) error {
	// 获取临时订单
	tempID, err := svc.orderRepo.GetTempOrder(ctx, userID)
	if err != nil || tempID != giftID {
		slog.Error("No Available Order")
		return errs.ErrNotFound
	}

	// 正式订单落库
	order := &model2.Order{
		ID:     svc.idGen.NextID(),
		UserID: userID,
		GiftID: giftID,
		Count:  1,
	}

	if err := svc.orderRepo.CreateOrder(ctx, order); err != nil {
		_ = svc.giftRepo.IncreaseCacheInventory(ctx, giftID)
		if errors.Is(err, repository2.ErrUniqueKey) {
			slog.Error("Server Internal Error", "error", err)
			return errs.ErrInternal
		}
		slog.Error("Server Internal Error", "error", err)
		return errs.ErrInternal
	}

	// 删除临时订单
	_ = svc.orderRepo.DeleteTempOrder(ctx, userID)
	return nil
}

func (svc *lotteryService) GiveUp(ctx context.Context, userID int64, giftID int64) error {
	// 获取临时订单
	tempID, err := svc.orderRepo.GetTempOrder(ctx, userID)
	if err != nil || tempID != giftID {
		slog.Error("No Available Order")
		return errs.ErrNotFound
	}

	_ = svc.orderRepo.DeleteTempOrder(ctx, userID)
	_ = svc.giftRepo.IncreaseCacheInventory(ctx, giftID)
	return nil
}

func (svc *lotteryService) Result(ctx context.Context, userID int64) (*model2.Order, *model2.Gift, error) {
	// 获取订单
	order, err := svc.orderRepo.GetOrder(ctx, userID)
	if err != nil {
		slog.Error("No Available Order")
		return nil, nil, errs.ErrNotFound
	}

	// 获取 Gift
	gift, err := svc.giftRepo.GetByID(ctx, order.GiftID)
	if err != nil {
		gift = &model2.Gift{}
	}

	return order, gift, nil
}

func (svc *lotteryService) StartLotteryOrderConsumer(ctx context.Context) {
	consumer := svc.mq.RocketConsumer
	for {
		select {
		case <-ctx.Done():
			slog.Info("关闭 Session Register Consumer 成功 ...")
			return
		default:
			messages, err := consumer.Receive(ctx, 1, conf.RocketLotteryInvisibleDuration) // 一批一条
			if err != nil {
				// 判断是否 broker 里暂时没有数据, 40401
				var e *rmq_client.ErrRpcStatus
				if errors.As(err, &e) && e.Code != 40401 {
					log.Printf("Receive Message Failed, Code %d, Error %s\n", e.Code, e.Message)
				}
				continue
			}

			for _, message := range messages {
				var order model2.Order
				err := sonic.Unmarshal(message.GetBody(), &order)
				if err != nil {
					continue
				}
				gid, _ := svc.orderRepo.GetTempOrder(ctx, order.UserID)
				if gid == order.GiftID {
					// 支付超时，删除临时订单，增加库存
					_ = svc.orderRepo.DeleteTempOrder(ctx, order.UserID)
					_ = svc.giftRepo.IncreaseCacheInventory(ctx, gid)
				}
				consumer.Ack(ctx, message)
			}
		}
	}
}

// InitCacheInventory 初始化库存
func (svc *lotteryService) InitCacheInventory(ctx context.Context) {
	svc.giftRepo.InitCacheInventory(ctx)
}

func (svc *lotteryService) produce(ctx context.Context, order *model2.Order, delay int) error {
	// 序列化 Order
	body, err := sonic.Marshal(order)
	if err != nil {
		return errs.ErrInternal
	}

	// 构造 Message
	message := &rmq_client.Message{
		Topic: conf.RocketLotteryTopic,
		Body:  body,
	}
	message.SetDelayTimestamp(time.Now().Add(time.Duration(delay) * time.Second))

	// 发送消息
	if _, err = svc.mq.RocketProducer.Send(ctx, message); err != nil {
		return errs.ErrInternal
	}
	return nil
}

// 抽奖算法
func lottery(probs []float64) int {
	if len(probs) == 0 {
		return -1
	}

	sum := 0.0
	acc := make([]float64, len(probs))
	for i, prob := range probs {
		sum += prob
		acc[i] = sum
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
