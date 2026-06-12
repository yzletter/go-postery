package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/yzletter/go-postery/backend/conf"
	"github.com/yzletter/go-postery/backend/event"
	"gorm.io/gorm"
)

func ScanOutbox(ctx context.Context, db *gorm.DB, producer *kafka.Writer) {
	interval := conf.OutboxInterval
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// 扫表失败退避时间
	backoff := time.Second

	for {
		select {
		case <-ctx.Done():
			slog.Info("close scan outbox success")
			return
		case <-ticker.C:
			// 加锁
			lockOwner := uuid.New().String() // 本轮加锁者
			updates := map[string]any{
				"status":       event.StatusEventProcessing,                                          // 发送中
				"lock_owner":   lockOwner,                                                            // 加锁者
				"locked_until": gorm.Expr("DATE_ADD(NOW(), INTERVAL ? SECOND)", conf.OutboxLockTime), // 锁过期时间
			}
			// 没有发送过, 或需要重试
			pendingOrRetry := db.
				Where("status IN (?, ?)", event.StatusEventPending, event.StatusEventRetry).
				Where("next_retry_at IS NULL OR next_retry_at <= NOW()").
				Where("lock_owner IS NULL OR locked_until IS NULL OR locked_until <= NOW()")
			// 发送中, 但锁已过期
			processingExpired := db.Where("status = ? AND (locked_until IS NULL OR locked_until <= NOW())", event.StatusEventProcessing)
			result := db.Model(&event.Event{}).
				Where(pendingOrRetry).
				Or(processingExpired).
				Order("created_at ASC").
				Limit(100).Updates(updates)
			if result.Error != nil || result.RowsAffected == 0 {
				slog.Error("this round lock failed or scan nothing", "error", result.Error)
				continue
			}

			// 查哪些 Event 抢到了锁
			var events []*event.Event
			result = db.Model(&event.Event{}).
				Where("status = ? AND lock_owner = ?", event.StatusEventProcessing, lockOwner).
				Order("created_at ASC").
				Limit(100).Find(&events)

			// 扫表失败
			if result.Error != nil {
				slog.Error("get locked events failed", "error", result.Error)
				select {
				case <-ctx.Done():
					slog.Info("close scan outbox success")
					return
				case <-time.After(backoff):
				}

				if backoff < 4*time.Second {
					backoff *= 2
				}
				continue
			}

			// 扫表成功重置退避
			backoff = time.Second

			// Kafka 发消息
			for _, e := range events {
				// 发送消息
				err := producer.WriteMessages(ctx, kafka.Message{
					Topic: e.Topic,
					Key:   []byte(e.MessageKey), Value: []byte(e.MessageValue),
				})

				// 发送失败回填表
				if err != nil {
					slog.Error("kafka send message failed", "error", err, "id", e.ID, "topic", e.Topic, "message", e.MessageKey, "value", e.MessageValue)
					status := event.StatusEventRetry
					if e.RetryCnt >= 5 {
						status = event.StatusEventFailed
					}
					updates := map[string]any{
						"status":        status,
						"next_retry_at": gorm.Expr("DATE_ADD(NOW(), INTERVAL ? SECOND)", 30),
						"retry_cnt":     gorm.Expr("retry_cnt + ?", 1),
						"lock_owner":    nil,
						"locked_until":  nil,
					}

					// 发送失败的更新失败错误可忽略，原消息在数据库的锁会释放
					result = db.Model(&event.Event{}).Where("id = ? AND lock_owner = ?", e.ID, lockOwner).Updates(updates)
					if result.Error != nil {
						slog.Error("update event failed", "error", result.Error, "id", e.ID, "topic", e.Topic, "message", e.MessageKey, "value", e.MessageValue)
					}
					continue
				}

				// 发送成功回填表释放锁
				updates := map[string]any{
					"status":       event.StatusEventSent,
					"lock_owner":   nil,
					"locked_until": nil,
				}
				result = db.Model(&event.Event{}).Where("id = ? AND lock_owner = ?", e.ID, lockOwner).Updates(updates)
				if result.Error != nil {
					slog.Error("update event failed", "error", result.Error, "id", e.ID, "topic", e.Topic, "message", e.MessageKey, "value", e.MessageValue)
				}
			}
		}
	}
}
