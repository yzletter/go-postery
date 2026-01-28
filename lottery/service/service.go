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
	lottery_grpc "github.com/yzletter/go-postery/api/proto/lottery/v1"
	"github.com/yzletter/go-postery/errno"
	"github.com/yzletter/go-postery/lottery/conf"
	"github.com/yzletter/go-postery/lottery/dto"
	infraRocketMQ "github.com/yzletter/go-postery/lottery/infra/rocketmq"
	"github.com/yzletter/go-postery/lottery/model"
	"github.com/yzletter/go-postery/lottery/repository"
	"github.com/yzletter/go-postery/lottery/service/ports"
)

type lotteryService struct {
	giftRepo  repository.GiftRepository
	orderRepo repository.OrderRepository
	mq        *infraRocketMQ.RocketMQ
	idGen     ports.IDGenerator
	lottery_grpc.UnimplementedLotteryServiceServer
}

func NewLotteryService(orderRepo repository.OrderRepository, giftRepo repository.GiftRepository, mq *infraRocketMQ.RocketMQ, idGen ports.IDGenerator) LotteryService {
	return &lotteryService{
		orderRepo:                         orderRepo,
		giftRepo:                          giftRepo,
		idGen:                             idGen,
		mq:                                mq,
		UnimplementedLotteryServiceServer: lottery_grpc.UnimplementedLotteryServiceServer{},
	}
}

func (svc *lotteryService) GetAllGifts(ctx context.Context, req *lottery_grpc.EmptyRequest) (*lottery_grpc.Gifts, error) {
	var empty = new(lottery_grpc.Gifts)
	gifts, err := svc.giftRepo.GetAllGifts(ctx)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return empty, errno.ErrGiftNotFound
		}
		return empty, errno.ErrServerInternal
	}

	respGifts := make([]*lottery_grpc.Gift, 0, len(gifts))
	for _, gift := range gifts {
		giftDTO := dto.ToGift(gift)
		respGifts = append(respGifts, giftDTO)
	}

	return &lottery_grpc.Gifts{Gifts: respGifts}, nil
}

func (svc *lotteryService) Lottery(ctx context.Context, id *lottery_grpc.UserID) (*lottery_grpc.Gift, error) {
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
			empty := &model.Gift{
				ID:   0,
				Name: "奖品已抽完",
			}
			return dto.ToGift(empty), nil
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
			return dto.ToGift(gift), nil
		}

		// 创建临时订单
		if err := svc.orderRepo.CreateTempOrder(ctx, id.UserID, gid); err != nil {
			_ = svc.giftRepo.IncreaseCacheInventory(ctx, gid)
			continue
		}

		// 发送延迟消息
		err = svc.produce(ctx, &model.Order{UserID: id.UserID, GiftID: gid}, conf.RocketLotteryPayDelay)
		if err != nil {
			_ = svc.giftRepo.IncreaseCacheInventory(ctx, gid)
			_ = svc.orderRepo.DeleteTempOrder(ctx, id.UserID)
			continue
		}

		// 返回数据
		return dto.ToGift(gift), nil
	}

	empty := &model.Gift{
		ID:   1,
		Name: "谢谢参与",
	}
	slog.Error("Lottery Nothing")
	return dto.ToGift(empty), nil
}

func (svc *lotteryService) Pay(ctx context.Context, req *lottery_grpc.LotteryCommonRequest) (*lottery_grpc.EmptyResponse, error) {
	// 获取临时订单
	tempID, err := svc.orderRepo.GetTempOrder(ctx, req.UserID)
	if err != nil || tempID != req.GiftID {
		return &lottery_grpc.EmptyResponse{}, errno.ErrNotLottery
	}

	// 正式订单落库
	order := &model.Order{
		ID:     svc.idGen.NextID(),
		UserID: req.UserID,
		GiftID: req.GiftID,
		Count:  1,
	}

	if err := svc.orderRepo.CreateOrder(ctx, order); err != nil {
		_ = svc.giftRepo.IncreaseCacheInventory(ctx, req.GiftID)
		if errors.Is(err, repository.ErrUniqueKey) {
			slog.Error("Create Order Failed", "error", err)
			return &lottery_grpc.EmptyResponse{}, errno.ErrServerInternal
		}
		return &lottery_grpc.EmptyResponse{}, errno.ErrServerInternal
	}

	// 删除临时订单
	_ = svc.orderRepo.DeleteTempOrder(ctx, req.UserID)
	return &lottery_grpc.EmptyResponse{}, nil
}

func (svc *lotteryService) GiveUp(ctx context.Context, req *lottery_grpc.LotteryCommonRequest) (*lottery_grpc.EmptyResponse, error) {
	// 获取临时订单
	tempID, err := svc.orderRepo.GetTempOrder(ctx, req.UserID)
	if err != nil || tempID != req.GiftID {
		return &lottery_grpc.EmptyResponse{}, errno.ErrNotLottery
	}

	_ = svc.orderRepo.DeleteTempOrder(ctx, req.UserID)
	_ = svc.giftRepo.IncreaseCacheInventory(ctx, req.GiftID)
	return &lottery_grpc.EmptyResponse{}, nil
}

func (svc *lotteryService) Result(ctx context.Context, id *lottery_grpc.UserID) (*lottery_grpc.Order, error) {
	var empty = new(lottery_grpc.Order)

	// 获取订单
	order, err := svc.orderRepo.GetOrder(ctx, id.UserID)
	if err != nil {
		return empty, errno.ErrOrderNotFound
	}

	// 获取 Gift
	gift, err := svc.giftRepo.GetByID(ctx, order.GiftID)
	if err != nil {
		gift = &model.Gift{}
	}

	return dto.ToOrder(order, gift), nil
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

// InitCacheInventory 初始化库存
func (svc *lotteryService) InitCacheInventory(ctx context.Context) {
	svc.giftRepo.InitCacheInventory(ctx)
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
	message.SetDelayTimestamp(time.Now().Add(time.Duration(delay) * time.Second))

	// 发送消息
	if _, err = svc.mq.RocketProducer.Send(ctx, message); err != nil {
		return errno.ErrServerInternal
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
