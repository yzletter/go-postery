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
	"github.com/yzletter/go-postery/microservice-backend/lottery/model"
	"github.com/yzletter/go-postery/microservice-backend/lottery/repository"
	"github.com/yzletter/go-postery/microservice-backend/lottery/service/ports"
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
		idGen:     idGen,
		mq:        mq,
	}
}

// InitCacheInventory 初始化库存
func (svc *lotteryService) InitCacheInventory(ctx context.Context) {
	svc.giftRepo.InitCacheInventory(ctx)
}

// GetAllGifts 获取所有礼物
func (svc *lotteryService) GetAllGifts(ctx context.Context) ([]*model.Gift, error) {
	gifts, err := svc.giftRepo.GetAllGifts(ctx)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			slog.Error("Gift Not Found", "error", err)
			return nil, errs.ErrNotFound
		}
		slog.Error("Server Internal Error", "error", err)
		return nil, errs.ErrInternal
	}

	return gifts, nil
}

// Lottery 抽奖接口
func (svc *lotteryService) Lottery(ctx context.Context, userID int64) (*model.Gift, error) {
	// 获取临时订单
	exists, err := svc.orderRepo.CheckTempOrder(ctx, userID)
	if err != nil {
		slog.Error("Check Temp Order Failed", "error", err)
		return nil, errs.ErrInternal
	} else if exists {
		empty := &model.Gift{
			ID:   2,
			Name: "当前有未支付的订单",
		}
		return empty, nil
	}

	// 进行十次抽奖尝试
	const maxTry = 10

	for try := 1; try <= maxTry; try++ {
		// 获取商品库存
		gifts, err := svc.giftRepo.GetCacheInventory(ctx)
		if err != nil {
			continue
		}

		// 获取商品 ID 和库存
		giftIDs, giftStocks := make([]int64, 0, len(gifts)), make([]float64, 0, len(gifts))
		for _, gift := range gifts {
			// 只保留 > 0 的部分
			if gift.Count > 0 {
				giftIDs = append(giftIDs, gift.ID)
				giftStocks = append(giftStocks, float64(gift.Count))
			}
		}

		// 所有奖品已抽完
		if len(giftStocks) == 0 {
			empty := &model.Gift{
				ID:   0,
				Name: "奖品已抽完",
			}
			return empty, nil
		}

		// 进行抽奖
		gid := lottery(giftIDs, giftStocks)
		if gid == -1 {
			continue
		}

		// 扣减库存
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
			if !errors.Is(err, repository.ErrResourceConflict) {
				slog.Error("Create Temp Order Failed", "error", err)
				return nil, errs.ErrInternal
			}
			empty := &model.Gift{
				ID:   2,
				Name: "当前有未支付的订单",
			}
			return empty, nil
		}

		// 发送延迟消息
		if err = svc.produce(ctx, &model.Order{UserID: userID, GiftID: gid}, conf.RocketLotteryPayDelay); err != nil {
			_ = svc.giftRepo.IncreaseCacheInventory(ctx, gid)
			_ = svc.orderRepo.DeleteTempOrder(ctx, userID)
			continue
		}

		// 返回数据
		return gift, nil
	}

	empty := &model.Gift{
		ID:   1,
		Name: "谢谢参与",
	}
	slog.Error("Lottery Nothing")
	return empty, nil
}

// Pay 支付接口
func (svc *lotteryService) Pay(ctx context.Context, userID int64, giftID int64) error {
	// 获取临时订单
	tempID, err := svc.orderRepo.GetTempOrder(ctx, userID)
	if err != nil || tempID != giftID {
		slog.Error("No Available Order")
		return errs.ErrNotFound
	}

	// 正式订单落库
	order := &model.Order{
		ID:     svc.idGen.NextID(),
		UserID: userID,
		GiftID: giftID,
		Count:  1,
	}

	if err := svc.orderRepo.CreateOrder(ctx, order); err != nil {
		_ = svc.giftRepo.IncreaseCacheInventory(ctx, giftID)
		if errors.Is(err, repository.ErrUniqueKey) {
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

// GiveUp 放弃支付
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

// Result 支付结果
func (svc *lotteryService) Result(ctx context.Context, userID int64) (*model.Order, *model.Gift, error) {
	// 获取订单
	order, err := svc.orderRepo.GetOrder(ctx, userID)
	if err != nil {
		slog.Error("No Available Order")
		return nil, nil, errs.ErrNotFound
	}

	// 获取 Gift
	gift, err := svc.giftRepo.GetByID(ctx, order.GiftID)
	if err != nil {
		gift = &model.Gift{}
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
				var order model.Order
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

func (svc *lotteryService) produce(ctx context.Context, order *model.Order, delay int) error {
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
func lottery(ids []int64, stocks []float64) int64 {
	if len(ids) == 0 || len(ids) != len(stocks) {
		return -1
	}

	sum := 0.0
	acc := make([]float64, len(stocks))
	for i, prob := range stocks {
		sum += prob
		acc[i] = sum
	}

	// 获取 [0, sum) 的随机数
	x := rand.Float64() * sum

	// 二分查找大于等于 x 的第一个数的位置
	l, r := 0, len(stocks)-1
	for l < r {
		mid := (l + r) / 2
		if acc[mid] < x {
			l = mid + 1
		} else {
			r = mid
		}
	}

	return ids[l]
}
