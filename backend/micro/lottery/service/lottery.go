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

/*
处理抽奖相关接口
*/

// todo 围绕用于区分抽奖场次的 ID 设计配套服务

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
func (svc *lotteryService) Lottery(ctx context.Context, userID int64, requestID int64) (*dto.LotteryResult, error) {
	const maxTry = 10 // 进行十次抽奖尝试
	backup := requestID
	for try := 1; try <= maxTry; try++ {
		requestID = backup

		// 获取预热好的商品
		gifts, err := svc.giftRepo.GetCacheInventory(ctx)
		if err != nil {
			slog.Error("get cache inventory failed", "error", err)
			return nil, errs.ErrInternal
		}

		// 记录商品 ID 和库存
		giftIDs, giftStocks := make([]int64, 0, len(gifts)), make([]float64, 0, len(gifts))
		for _, gift := range gifts {
			// 只保留 > 0 的部分
			if gift.Count > 0 {
				giftIDs = append(giftIDs, gift.ID)
				giftStocks = append(giftStocks, float64(gift.Count))
			}
		}

		giftID := int64(0)
		needLottery := true

		// 所有奖品已抽完
		if len(giftStocks) == 0 {
			// 若当前请求为并发请求, 尝试复用其他请求走后续流程
			requestID, giftID, err = svc.giftRepo.Reuse(ctx, requestID, userID)
			if err != nil {
				// 明确没有可复用的预占记录, 结束本次抽奖
				if errors.Is(err, repository.ErrRecordNotFound) {
					break
				}
				// 查询结果可能未知, 不能视为未中奖
				slog.Error("reuse lottery request failed", "user_id", userID, "request_id", requestID, "error", err)
				return nil, errs.ErrUnavailable
			}

			needLottery = false
		}

		if needLottery {
			// 计算选出的 ID
			giftID = lottery(giftIDs, giftStocks)
			if giftID == -1 {
				continue
			}

			// 尝试进行抽奖: 验证抽奖资格 + 扣减库存 + 保存扣减记录
			requestID, giftID, err = svc.giftRepo.Lottery(ctx, requestID, userID, giftID)
			switch {
			case err == nil:
				// 扣减成功, 继续创建订单
			case errors.Is(err, repository.ErrInsufficientStock):
				// 明确没有扣成功, 进入下一轮循环
				continue
			case errors.Is(err, repository.ErrResourceConflict):
				// 预扣过, 复用
			default:
				// 其他错误
				return nil, errs.ErrUnavailable // TODO 区分没有抽奖资格和其他错误
			}
		}

		// 构造订单结构体
		lotteryTime := time.Now() // 抽奖时间
		order := &model.Order{
			ID:                  svc.idGen.NextID(),
			UserID:              userID,
			GiftID:              giftID,
			Count:               1,
			ExpireAt:            lotteryTime.Add(conf.RocketLotteryPayDelay * time.Second), // 支付超时时间
			Status:              model.OrderStatusPending,                                  // 订单待支付
			StockRollbackStatus: model.StockRollbackStatusPending,                          // 库存待回补
		}

		// 尝试订单落库并返回, 若已存在订单会直接返回
		savedOrder, err := svc.orderRepo.CreateOrder(ctx, order)
		if err != nil || savedOrder == nil {
			slog.Error("create order failed", "user_id", userID, "request_id", requestID, "error", err)
			return nil, errs.ErrInternal
		}

		// 只有支付期限内的待支付订单可以继续返回中奖结果
		if savedOrder.Status == model.OrderStatusPaid {
			return nil, errs.ErrAlreadyExits
		}
		if savedOrder.Status != model.OrderStatusPending || !savedOrder.ExpireAt.After(time.Now()) {
			return nil, errs.ErrNotFound
		}

		// 尝试发送延迟消息
		if err = svc.produce(ctx, savedOrder, conf.RocketLotteryPayDelay); err != nil {
			// 延迟消息发送失败, 导致无法通过 RocketMQ 的延迟消息回收库存
			// 最多会导致库存流失, 但不会超卖, 通过定时扫订单表进行兜底
			slog.Error("produce temp order failed", "order_id", savedOrder.ID, "user_id", userID, "gift_id", savedOrder.GiftID, "error", err)
		}

		// 获取商品详情, 这里应该用返回的 order 的 giftID
		gift, err := svc.giftRepo.GetByID(ctx, savedOrder.GiftID)
		if err != nil {
			// 降级
			slog.Warn("get gift detail failed", "gift_id", giftID, "error", err)
			gift = &model.Gift{
				ID:   savedOrder.GiftID,
				Name: "未知商品",
			}
		}

		// 返回数据
		return &dto.LotteryResult{
			Success:     true,
			Description: dto.DescriptionLotterySuccess,
			OrderID:     savedOrder.ID,
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
	// 获取订单，包含已支付订单
	order, err := svc.orderRepo.GetOrder(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return errs.ErrNotFound
		}
		slog.Error("get order before pay failed", "user_id", userID, "error", err)
		return errs.ErrInternal
	}

	// 校验订单归属和请求参数
	if order == nil || order.UserID != userID || order.ID != tempOrderID || order.GiftID != giftID {
		slog.Info("pay rejected: no available order", "user_id", userID, "temp_order_id", tempOrderID, "gift_id", giftID)
		return errs.ErrNotFound
	}

	// 支付重试幂等：已经支付过，直接成功
	if order.Status == model.OrderStatusPaid {
		return nil
	}

	// todo 接支付链路

	// MySQL 事务内校验支付状态和有效期，更新订单状态并扣减实际库存
	if err := svc.orderRepo.Pay(ctx, order.ID); err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return errs.ErrNotFound
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

	// 取消临时订单 Pending -> Cancelled
	if err = svc.orderRepo.CancelTempOrder(ctx, tempOrder.ID); err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			slog.Info("give up skipped: order not found", "user_id", userID, "temp_order_id", tempOrderID)
			return errs.ErrNotFound
		}
		slog.Error("cancel temp order failed", "user_id", userID, "temp_order_id", tempOrderID, "error", err)
		return errs.ErrInternal
	}

	if !svc.rollbackStock(ctx, tempOrder.ID, tempOrder.GiftID) {
		slog.Info("roll back failed", "temp_order_id", tempOrder.ID)
		return nil
	}

	return nil
}

// Result 支付结果
func (svc *lotteryService) Result(ctx context.Context, userID int64) (*model.Order, *model.Gift, error) {
	// 获取订单
	order, err := svc.orderRepo.GetOrder(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			slog.Info("lottery result not found", "user_id", userID)
			return nil, nil, errs.ErrNotFound
		}
		slog.Error("get lottery result failed", "user_id", userID, "error", err)
		return nil, nil, errs.ErrInternal
	}

	// 获取 Gift
	gift, err := svc.giftRepo.GetByID(ctx, order.GiftID)
	if err != nil {
		gift = &model.Gift{}
	}

	return order, gift, nil
}

// 抽奖算法
func lottery(ids []int64, stocks []float64) int64 {
	if len(ids) == 0 || len(ids) != len(stocks) {
		return -1
	}

	// 累加概率
	sum, acc := 0.0, make([]float64, len(stocks))
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

// 发送延迟消息
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

// 获取临时订单
func (svc *lotteryService) getTempOrder(ctx context.Context, userID int64) (*dto.LotteryResult, bool, error) {
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
			Name: "未知商品",
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
