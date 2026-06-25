package service

import (
	"context"
	"errors"
	"log/slog"
	"math/rand"
	"time"

	rmq_client "github.com/apache/rocketmq-clients/golang/v5"
	"github.com/bytedance/sonic"
	"github.com/yzletter/go-postery/backend/conf"
	"github.com/yzletter/go-postery/backend/grpc/errs"
	infraRocketMQ "github.com/yzletter/go-postery/backend/infra/mq/rocketmq"
	"github.com/yzletter/go-postery/backend/micro/lottery/dto"
	"github.com/yzletter/go-postery/backend/micro/lottery/model"
	"github.com/yzletter/go-postery/backend/micro/lottery/repository"
	"github.com/yzletter/go-postery/backend/ports"
)

type lotteryService struct {
	giftRepo  repository.GiftRepository
	orderRepo repository.OrderRepository
	mq        *infraRocketMQ.RocketMQ
	idGen     ports.IDGenerator
}

const (
	stockRollbackRetryDelay        = 10 * time.Second
	stockRollbackScannerPeriod     = 10 * time.Second
	stockRollbackScanBatchLimit    = 100
	lotteryOrderConsumerBatchLimit = 10
)

func NewLotteryService(orderRepo repository.OrderRepository, giftRepo repository.GiftRepository, mq *infraRocketMQ.RocketMQ, idGen ports.IDGenerator) LotteryService {
	return &lotteryService{
		orderRepo: orderRepo,
		giftRepo:  giftRepo,
		idGen:     idGen,
		mq:        mq,
	}
}

// InitCacheInventory 初始化抽奖库存 todo 用于区分每次抽奖的 ID
func (svc *lotteryService) InitCacheInventory(ctx context.Context) {
	svc.giftRepo.InitCacheInventory(ctx)
}

// GetAllGifts 获取所有抽奖礼物
func (svc *lotteryService) GetAllGifts(ctx context.Context) ([]*model.Gift, error) {
	gifts, err := svc.giftRepo.GetAllGifts(ctx)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			slog.Info("gift list empty")
			return nil, errs.ErrNotFound
		}
		slog.Error("get all gifts failed", "error", err)
		return nil, errs.ErrInternal
	}

	return gifts, nil
}

// Lottery 抽奖接口
func (svc *lotteryService) Lottery(ctx context.Context, userID int64) (*dto.LotteryResult, error) {
	// 尝试获取临时订单查看是否抽过奖
	if result, exists, err := svc.getTempOrderResult(ctx, userID); err != nil {
		slog.Error("get temp order result failed", "user_id", userID, "error", err)
		return nil, errs.ErrInternal
	} else if exists {
		return result, nil
	}

	// 进行十次抽奖尝试
	const maxTry = 10

	for try := 1; try <= maxTry; try++ {
		// 获取商品库存
		gifts, err := svc.giftRepo.GetCacheInventory(ctx)
		if err != nil {
			slog.Error("get cache inventory failed", "error", err)
			return nil, errs.ErrInternal
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
			return &dto.LotteryResult{
				Success:     false,
				Description: dto.DescriptionNoGifts,
				UserID:      userID,
			}, nil
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
			slog.Warn("get gift after stock reduce failed", "gift_id", gid, "error", err)
			_ = svc.giftRepo.IncreaseCacheInventory(ctx, gid)
			continue
		}

		lotteryTime := time.Now() // 抽奖时间
		order := &model.Order{
			ID:       svc.idGen.NextID(),
			UserID:   userID,
			GiftID:   gid,
			Count:    1,
			ExpireAt: lotteryTime.Add(conf.RocketLotteryPayDelay * time.Second),
		}

		// 创建临时订单
		if err := svc.orderRepo.CreateTempOrder(ctx, order); err != nil {
			_ = svc.giftRepo.IncreaseCacheInventory(ctx, gid)
			if !errors.Is(err, repository.ErrUniqueKey) {
				// 如果不是订单重复, 报错
				slog.Error("create temp order failed", "user_id", userID, "gift_id", gid, "error", err)
				return nil, errs.ErrInternal
			}

			//订单重复，获取已经存在的临时订单进行返回
			if result, exists, err := svc.getTempOrderResult(ctx, userID); err != nil {
				slog.Error("get temp order result failed", "user_id", userID, "error", err)
				return nil, errs.ErrInternal
			} else if exists {
				return result, nil
			}

			continue
		}

		// 发送延迟消息
		if err = svc.produce(ctx, order, conf.RocketLotteryPayDelay); err != nil {
			// 延迟消息发送失败, 最多导致超时订单无法回收库存，导致库存流失, 但不会超卖
			// 定时扫订单表进行兜底
			slog.Error("produce temp order failed", "order_id", order.ID, "user_id", userID, "gift_id", gid, "error", err)
		}

		// 返回数据
		return &dto.LotteryResult{
			Success:     true,
			Description: dto.DescriptionLotterySuccess,
			OrderID:     order.ID,
			UserID:      userID,
			Gift:        gift,
		}, nil
	}

	// 未抽中奖品
	slog.Info("lottery result empty", "user_id", userID)
	return &dto.LotteryResult{
		Success:     true,
		Description: dto.DescriptionLotteryNothing,
		UserID:      userID,
	}, nil
}

// Pay 支付接口
func (svc *lotteryService) Pay(ctx context.Context, userID int64, tempOrderID int64, giftID int64) error {
	// 获取临时订单
	tempOrder, err := svc.orderRepo.GetTempOrder(ctx, userID)
	if err != nil || tempOrder == nil || tempOrder.ID != tempOrderID || tempOrder.GiftID != giftID {
		slog.Info("pay rejected: no available order", "user_id", userID, "temp_order_id", tempOrderID, "gift_id", giftID, "error", err)
		return errs.ErrNotFound
	}

	// todo 接支付链路

	// 正式订单落库 + 扣减实际库存
	if err := svc.orderRepo.PayTempOrder(ctx, tempOrder.ID); err != nil {
		if errors.Is(err, repository.ErrUniqueKey) {
			slog.Info("pay skipped: order already created", "user_id", userID, "temp_order_id", tempOrderID)
			return errs.ErrAlreadyExits
		}
		slog.Error("pay temp order failed", "user_id", userID, "temp_order_id", tempOrderID, "error", err)
		return errs.ErrInternal
	}

	return nil
}

// GiveUp 放弃支付
func (svc *lotteryService) GiveUp(ctx context.Context, userID int64, tempOrderID int64, giftID int64) error {
	// 获取临时订单
	tempOrder, err := svc.orderRepo.GetTempOrder(ctx, userID)
	if err != nil || tempOrder.ID != tempOrderID || tempOrder.GiftID != giftID {
		slog.Info("give up rejected: no available order", "user_id", userID, "temp_order_id", tempOrderID, "gift_id", giftID, "error", err)
		return errs.ErrNotFound
	}

	// 取消临时订单
	if err = svc.orderRepo.CancelTempOrder(ctx, tempOrder.ID); err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			slog.Info("give up skipped: order not found", "user_id", userID, "temp_order_id", tempOrderID)
			return errs.ErrNotFound
		}
		slog.Error("cancel temp order failed", "user_id", userID, "temp_order_id", tempOrderID, "error", err)
		return errs.ErrInternal
	}

	if !svc.rollbackOrderStock(ctx, tempOrder) {
		return errs.ErrInternal
	}

	return nil
}

// Result 支付结果
func (svc *lotteryService) Result(ctx context.Context, userID int64) (*model.Order, *model.Gift, error) {
	// 获取订单
	order, err := svc.orderRepo.GetOrder(ctx, userID)
	if err != nil {
		slog.Info("lottery result not found", "user_id", userID, "error", err)
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
			slog.Info("close lottery order consumer success")
			return
		default:
			messages, err := consumer.Receive(ctx, lotteryOrderConsumerBatchLimit, conf.RocketLotteryInvisibleDuration) // 一批一条
			if err != nil {
				// 判断是否 broker 里暂时没有数据, 40401
				var e *rmq_client.ErrRpcStatus
				if errors.As(err, &e) && e.Code != 40401 {
					slog.Warn("receive lottery order message failed", "code", e.Code, "error", e.Message)
				}
				continue
			}

			for _, message := range messages {
				var tempOrder model.Order
				if err := sonic.Unmarshal(message.GetBody(), &tempOrder); err != nil {
					// 毒消息进行 Ack 避免恶性循环
					slog.Warn("invalid lottery temp order message, skip", "error", err, "message_id", message.GetMessageId(), "topic", message.GetTopic())
					// 毒消息 Ack 失败
					if ackErr := consumer.Ack(ctx, message); ackErr != nil {
						slog.Error("ack invalid lottery message failed", "error", ackErr, "message_id", message.GetMessageId())
					}
					continue
				}

				if needRollback, err := svc.orderRepo.RecycleTempOrder(ctx, tempOrder.UserID, tempOrder.ID); err != nil {
					slog.Error("recycle temp order failed", "error", err, "order_id", tempOrder.ID, "user_id", tempOrder.UserID)
					continue
				} else if needRollback && !svc.rollbackOrderStock(ctx, &tempOrder) {
					continue
				}

				// 消息进行 ACK
				if ackErr := consumer.Ack(ctx, message); ackErr != nil {
					slog.Error("ack lottery message failed", "error", ackErr, "message_id", message.GetMessageId())
				}

			}
		}
	}
}

// StartStockRollbackScanner 扫描失败库存回补
func (svc *lotteryService) StartStockRollbackScanner(ctx context.Context) {
	ticker := time.NewTicker(stockRollbackScannerPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("close stock rollback scanner success")
			return
		case <-ticker.C:
			orders, err := svc.orderRepo.ListRollbackDueOrders(ctx, stockRollbackScanBatchLimit)
			if err != nil {
				slog.Error("list stock rollback orders failed", "error", err)
				continue
			}

			for _, order := range orders {
				if order == nil {
					continue
				}

				// 是否需要回补
				if needRollback, err := svc.orderRepo.RecycleTempOrder(ctx, order.UserID, order.ID); err != nil {
					slog.Error("recycle rollback due order failed", "error", err, "order_id", order.ID, "user_id", order.UserID)
					continue
				} else if needRollback {
					_ = svc.rollbackOrderStock(ctx, order)
				}
			}
		}
	}
}

// rollbackOrderStock 进行库存回补并返回是否成功
func (svc *lotteryService) rollbackOrderStock(ctx context.Context, order *model.Order) bool {
	if err := svc.giftRepo.RollbackCacheInventory(ctx, order.ID, order.GiftID); err != nil {
		slog.Error("rollback gift cache inventory failed", "error", err, "order_id", order.ID, "gift_id", order.GiftID)
		if markErr := svc.orderRepo.MarkRollbackFailed(ctx, order.ID, time.Now().Add(stockRollbackRetryDelay)); markErr != nil {
			slog.Error("mark stock rollback failed status failed", "error", markErr, "order_id", order.ID)
			return false
		}
		return true
	}

	if err := svc.orderRepo.MarkRollbackDone(ctx, order.ID); err != nil {
		slog.Error("mark stock rollback done failed", "error", err, "order_id", order.ID)
		return false
	}
	return true
}

func (svc *lotteryService) produce(ctx context.Context, order *model.Order, delay int) error {
	// 序列化 Order
	body, err := sonic.Marshal(order)
	if err != nil {
		return errs.ErrInternal
	}

	// 构造 Message
	message := &rmq_client.Message{Topic: conf.RocketLotteryTopic, Body: body}

	// 给 Message 添加延迟
	message.SetDelayTimestamp(time.Now().Add(time.Duration(delay) * time.Second))

	// 发送 Message
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

func (svc *lotteryService) getTempOrderResult(ctx context.Context, userID int64) (*dto.LotteryResult, bool, error) {
	tempOrder, err := svc.orderRepo.GetTempOrder(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return nil, false, nil
		}
		return nil, false, err
	}

	gift, err := svc.giftRepo.GetByID(ctx, tempOrder.GiftID)
	if err != nil {
		// 降级
		gift = &model.Gift{
			ID:   tempOrder.GiftID,
			Name: "奖品查询失败",
		}
	}

	return &dto.LotteryResult{
		Success:     false,
		Description: dto.DescriptionTempOrderToPay,
		OrderID:     tempOrder.ID,
		UserID:      tempOrder.UserID,
		Gift:        gift,
	}, true, nil
}
