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

			// 遍历 Message
			for _, message := range messages {
				var order model.Order
				// 反序列化消息
				if err := sonic.Unmarshal(message.GetBody(), &order); err != nil {
					// 毒消息进行 Ack 避免恶性循环
					slog.Warn("invalid lottery temp order message, skip", "error", err, "message_id", message.GetMessageId(), "topic", message.GetTopic())
					// 毒消息 Ack 失败
					if ackErr := consumer.Ack(ctx, message); ackErr != nil {
						slog.Error("ack invalid lottery message failed", "error", ackErr, "message_id", message.GetMessageId())
					}
					continue
				}

				// 进行库存回收
				// 情况 1. 需要回收
				//		情况 1.1 需要回收, 回收成功, ACK
				//		情况 1.2 需要回收, 回收失败
				// 			情况 1.2.1 需要回收, 回收失败, 但库存状态落库成功, 即订单状态为回补失败, ACK
				// 			情况 1.2.2 需要回收, 回收失败, 但库存状态落库失败, 即订单状态仍为待回补, 不 ACK
				// 情况 2. 不需要回收, ACK
				// 情况 3. 其他错误, 不 ACK
				if ack, err := svc.Recycle(ctx, order.UserID, order.ID, order.GiftID); err != nil {
					slog.Error("recycle temp order failed", "error", err, "order_id", order.ID, "user_id", order.UserID)
					// 发生错误
					continue
				} else if !ack {
					// 回补状态未成功更新
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

func (svc *lotteryService) Recycle(ctx context.Context, uid int64, orderID int64, giftID int64) (bool, error) {
	needRollback, err := svc.orderRepo.HandleTempOrder(ctx, uid, orderID)
	if err != nil {
		slog.Error("handle temp order failed", "error", err, "order_id", orderID, "user_id", uid)
		return false, err
	}

	if needRollback && !svc.rollbackStock(ctx, orderID, giftID) {
		// 需要回收但回收失败，不 ACK 消息
		return false, nil
	}

	return true, nil
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
				if needRollback, err := svc.orderRepo.HandleTempOrder(ctx, order.UserID, order.ID); err != nil {
					slog.Error("recycle rollback due order failed", "error", err, "order_id", order.ID, "user_id", order.UserID)
					continue
				} else if needRollback {
					_ = svc.rollbackStock(ctx, order.ID, order.GiftID)
				}
			}
		}
	}
}

// rollbackStock 进行库存回收并返回补偿状态是否更新成功
func (svc *lotteryService) rollbackStock(ctx context.Context, orderID int64, giftID int64) bool {
	if err := svc.giftRepo.RollbackCacheInventory(ctx, orderID, giftID); err != nil {
		// 库存回收失败
		slog.Error("rollback gift cache inventory failed", "error", err, "order_id", orderID, "gift_id", giftID)
		// 标记库存回补失败
		if markErr := svc.orderRepo.MarkRollbackFailed(ctx, orderID, time.Now().Add(stockRollbackRetryDelay)); markErr != nil {
			slog.Error("mark stock rollback failed status failed", "error", markErr, "order_id", orderID)
			return false // 库存回收失败, 库存回补状态也更新失败, 不能 ACK
		}

		return true // 库存回收失败, 库存回补状态更新成功，可以 ACK, 交由扫表回补
	}

	// 库存回补成功
	if err := svc.orderRepo.MarkRollbackDone(ctx, orderID); err != nil {
		slog.Error("mark stock rollback done failed", "error", err, "order_id", orderID)
		return false // 库存回补成功, 但是回补状态更新不成功，不能 ACK
	}

	// 库存回补成功, 回补状态更新也成功
	return true
}
