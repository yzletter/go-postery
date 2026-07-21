package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/yzletter/go-postery/backend/conf"
	"github.com/yzletter/go-postery/backend/event/outbox/model"
	"gorm.io/gorm"
)

const lockBatch = 100

func ScanOutbox(ctx context.Context, db *gorm.DB, producer *kafka.Writer) {
	interval := conf.OutboxInterval
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	logger := slog.With("component", "outbox_scanner")
	logger.Info("outbox scanner started", "interval", interval, "lock_batch", lockBatch)

	// 扫表失败退避时间
	backoff := time.Second

	for {
		select {
		case <-ctx.Done():
			logger.Info("outbox scanner stopped", "reason", ctx.Err())
			return
		case <-ticker.C:
			// 加锁
			lockOwner := uuid.New().String() // 本轮加锁者
			updates := map[string]any{
				"status":       model.OutboxEventStatusProcessing,                                    // 发送中
				"lock_owner":   lockOwner,                                                            // 加锁者
				"locked_until": gorm.Expr("DATE_ADD(NOW(), INTERVAL ? SECOND)", conf.OutboxLockTime), // 锁过期时间
			}

			// 第一部分：没有发送过, 或需要重试
			pendingOrRetry := db.
				Where("status IN (?, ?)", model.OutboxEventStatusPending, model.OutboxEventStatusRetry). // 等待发送 或 需要重试
				Where("next_retry_at IS NULL OR next_retry_at <= NOW()").
				Where("lock_owner IS NULL OR locked_until IS NULL OR locked_until <= NOW()")

			// 第二部分：发送中, 但锁已过期
			processingExpired := db.Where("status = ? AND (locked_until IS NULL OR locked_until <= NOW())", model.OutboxEventStatusProcessing)

			// 进行加锁
			result := db.Model(&model.OutboxEvent{}).
				Where(pendingOrRetry).Or(processingExpired).
				Order("created_at ASC").
				Limit(lockBatch).Updates(updates)
			if result.Error != nil {
				logger.Error("lock outbox events failed", "error", result.Error, "lock_owner", lockOwner)
				continue
			}
			if result.RowsAffected == 0 {
				logger.Debug("no outbox events ready")
				continue
			}
			lockedRows := result.RowsAffected

			// 查哪些 Event 抢到了锁
			var events []*model.OutboxEvent
			result = db.Model(&model.OutboxEvent{}).
				Where("status = ? AND lock_owner = ?", model.OutboxEventStatusProcessing, lockOwner).
				Order("created_at ASC").
				Limit(lockBatch).Find(&events)

			// 扫表失败
			if result.Error != nil {
				logger.Error("get locked outbox events failed", "error", result.Error, "lock_owner", lockOwner, "backoff", backoff)
				select {
				case <-ctx.Done():
					logger.Info("outbox scanner stopped", "reason", ctx.Err())
					return
				case <-time.After(backoff):
				}

				if backoff < 10*time.Second {
					backoff *= 2
				}
				continue
			}

			// 扫表成功重置退避
			backoff = time.Second

			if len(events) == 0 {
				logger.Warn("locked outbox events not found", "lock_owner", lockOwner, "locked_rows", lockedRows)
				continue
			}

			var sentCnt, retryCnt, failedCnt, updateFailedCnt int

			// Kafka 发消息
			for _, e := range events {
				// 发送消息
				err := producer.WriteMessages(ctx, kafka.Message{
					Topic: e.Topic,
					Key:   []byte(e.MessageKey), Value: []byte(e.MessageValue),
				})

				// 发送失败回填表释放锁
				if err != nil {
					status := model.OutboxEventStatusRetry
					// 超过五次重试的消息标记为失败
					if e.RetryCnt >= 5 {
						status = model.OutboxEventStatusFailed
						failedCnt++
					} else {
						retryCnt++
					}

					logger.Error(
						"send outbox event failed",
						"error", err,
						"event_id", e.ID,
						"topic", e.Topic,
						"message_key", e.MessageKey,
						"retry_cnt", e.RetryCnt,
						"next_status", outboxStatusName(status),
					)

					updates := map[string]any{
						"status":        status,
						"next_retry_at": gorm.Expr("DATE_ADD(NOW(), INTERVAL ? SECOND)", 15),
						"retry_cnt":     gorm.Expr("retry_cnt + ?", 1),
						"lock_owner":    nil,
						"locked_until":  nil,
					}

					// 发送失败的更新失败错误可忽略，原消息在数据库的锁会释放
					result = db.Model(&model.OutboxEvent{}).Where("id = ? AND lock_owner = ?", e.ID, lockOwner).Updates(updates)
					if result.Error != nil {
						updateFailedCnt++
						logger.Error(
							"mark outbox event retry failed",
							"error", result.Error,
							"event_id", e.ID,
							"topic", e.Topic,
							"message_key", e.MessageKey,
							"lock_owner", lockOwner,
							"next_status", outboxStatusName(status),
						)
					}
					continue
				}

				// 发送成功回填表释放锁
				updates := map[string]any{
					"status":       model.OutboxEventStatusSent,
					"lock_owner":   nil,
					"locked_until": nil,
				}
				result = db.Model(&model.OutboxEvent{}).Where("id = ? AND lock_owner = ?", e.ID, lockOwner).Updates(updates)
				if result.Error != nil {
					updateFailedCnt++
					logger.Error(
						"mark outbox event sent failed",
						"error", result.Error,
						"event_id", e.ID,
						"topic", e.Topic,
						"message_key", e.MessageKey,
						"lock_owner", lockOwner,
					)
					continue
				}
				sentCnt++
			}

			logger.Info(
				"outbox batch processed",
				"lock_owner", lockOwner,
				"locked", lockedRows,
				"loaded", len(events),
				"sent", sentCnt,
				"retry", retryCnt,
				"failed", failedCnt,
				"update_failed", updateFailedCnt,
			)
		}
	}
}

func outboxStatusName(status int) string {
	switch status {
	case model.OutboxEventStatusPending:
		return "pending"
	case model.OutboxEventStatusProcessing:
		return "processing"
	case model.OutboxEventStatusSent:
		return "sent"
	case model.OutboxEventStatusRetry:
		return "retry"
	case model.OutboxEventStatusFailed:
		return "failed"
	default:
		return "unknown"
	}
}
