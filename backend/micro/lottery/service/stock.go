package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	rmq_client "github.com/apache/rocketmq-clients/golang/v5"
	"github.com/bytedance/sonic"
	"github.com/yzletter/go-postery/backend/conf"
	"github.com/yzletter/go-postery/backend/micro/lottery/model"
)

/*
处理库存相关接口
*/

// todo 围绕用于区分抽奖场次的 ID 设计配套服务

const (
	stockRollbackRetryDelay        = 10 * time.Second // 回滚重试延迟
	stockRollbackScannerPeriod     = 10 * time.Second // 回滚扫描间隔
	stockRollbackScanBatchLimit    = 100              // 回滚失败订单扫描每批订单数
	lotteryOrderConsumerBatchLimit = 10               // 消息扫描每批消息数
)

// InitCacheInventory 初始化抽奖库存
func (svc *lotteryService) InitCacheInventory(ctx context.Context) {
	svc.giftRepo.InitCacheInventory(ctx)
}

// StartLotteryOrderConsumer 消费消息，进行库存回收
func (svc *lotteryService) StartLotteryOrderConsumer(ctx context.Context) {
	consumer := svc.mq.RocketConsumer
	for {
		select {
		case <-ctx.Done():
			slog.Info("close lottery order consumer success")
			return
		default:
			messages, err := consumer.Receive(ctx, lotteryOrderConsumerBatchLimit, conf.RocketLotteryInvisibleDuration)
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

				// 判断是否需要进行库存回收
				if needRollback, err := svc.orderRepo.RecycleTempOrder(ctx, tempOrder.UserID, tempOrder.ID); err != nil {
					slog.Error("recycle temp order failed", "error", err, "order_id", tempOrder.ID, "user_id", tempOrder.UserID)
					continue
				} else if needRollback && !svc.rollbackStock(ctx, &tempOrder) {
					// 需要回收但回收失败，不 ACK 消息
					continue
				}

				// 消息处理成功，进行 ACK
				if ackErr := consumer.Ack(ctx, message); ackErr != nil {
					slog.Error("ack lottery message failed", "error", ackErr, "message_id", message.GetMessageId())
				}

			}
		}
	}
}

// StartStockRollbackScanner 定时扫描库存回收失败，进行重试
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

				// 是否需要回收
				if needRollback, err := svc.orderRepo.RecycleTempOrder(ctx, order.UserID, order.ID); err != nil {
					slog.Error("recycle rollback due order failed", "error", err, "order_id", order.ID, "user_id", order.UserID)
					continue
				} else if needRollback {
					_ = svc.rollbackStock(ctx, order)
				}
			}
		}
	}
}

// rollbackStock 进行库存回收并返回补偿记录是否成功落库
func (svc *lotteryService) rollbackStock(ctx context.Context, order *model.Order) bool {
	if err := svc.giftRepo.RollbackCacheInventory(ctx, order.ID, order.GiftID); err != nil {
		// 库存回收失败
		slog.Error("rollback gift cache inventory failed", "error", err, "order_id", order.ID, "gift_id", order.GiftID)
		if markErr := svc.orderRepo.MarkRollbackFailed(ctx, order.ID, time.Now().Add(stockRollbackRetryDelay)); markErr != nil {
			slog.Error("mark stock rollback failed status failed", "error", markErr, "order_id", order.ID)
			return false // 库存回收失败, 记录落库也失败, 不能 ACK
		}

		return true // 库存回收失败, 但是记录落库成功，可以 ACK, 交由扫表回补
	}

	// 库存回补成功
	if err := svc.orderRepo.MarkRollbackDone(ctx, order.ID); err != nil {
		slog.Error("mark stock rollback done failed", "error", err, "order_id", order.ID)
		return false // 库存回补成功, 但是记录落库不成功，不能 ACK
	}

	// 库存回补成功, 记录落库也成功
	return true
}
