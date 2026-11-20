package service

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/yzletter/go-postery/backend/conf"
	"github.com/yzletter/go-postery/backend/micro/outbox/model"
	"gorm.io/gorm"
)

const (
	lockBatch    = 100
	batchTimeout = 40 * time.Second // 整轮处理必须在租约过期前结束
	sendTimeout  = 20 * time.Second // 给状态回填留出时间
)

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
			// Scanner 可能在停机信号和 ticker 同时就绪时选中 ticker, 所以抢占前再检查一次 ctx.Err(), 判断是否停机了
			if ctx.Err() != nil {
				return
			}

			// 收到停机信号时，cancel() 会连当前批次一起取消，Kafka 已确认的消息也可能来不及回填 sent
			// 所以用 context.WithoutCancel(ctx)
			batchCtx, cancelBatch := context.WithTimeout(context.WithoutCancel(ctx), batchTimeout)
			batchDB := db.WithContext(batchCtx)

			// 加锁
			lockOwner := uuid.New().String() // 本轮加锁者
			updates := map[string]any{
				"status":       model.OutboxEventStatusProcessing,                                    // 发送中
				"lock_owner":   lockOwner,                                                            // 加锁者
				"locked_until": gorm.Expr("DATE_ADD(NOW(), INTERVAL ? SECOND)", conf.OutboxLockTime), // 锁过期时间
			}

			// 第一部分：没有发送过, 或需要重试
			pendingOrRetry := batchDB.
				Where("status IN (?, ?)", model.OutboxEventStatusPending, model.OutboxEventStatusRetry). // 等待发送 或 需要重试
				Where("next_retry_at IS NULL OR next_retry_at <= NOW()").
				Where("lock_owner IS NULL OR locked_until IS NULL OR locked_until <= NOW()")

			// 第二部分：发送中, 但锁已过期
			processingExpired := batchDB.Where("status = ? AND (locked_until IS NULL OR locked_until <= NOW())", model.OutboxEventStatusProcessing)

			// 进行加锁
			result := batchDB.Model(&model.OutboxEvent{}).
				Where(pendingOrRetry).Or(processingExpired).
				Order("created_at ASC").
				Limit(lockBatch).Updates(updates)
			if result.Error != nil {
				cancelBatch()
				logger.Error("lock outbox events failed", "error", result.Error, "lock_owner", lockOwner)
				continue
			}
			if result.RowsAffected == 0 {
				cancelBatch()
				logger.Debug("no outbox events ready")
				continue
			}
			lockedRows := result.RowsAffected

			// 查哪些 Event 抢到了锁
			var events []*model.OutboxEvent
			result = batchDB.Model(&model.OutboxEvent{}).
				Where("status = ? AND lock_owner = ? AND locked_until > NOW()", model.OutboxEventStatusProcessing, lockOwner).
				Order("created_at ASC").
				Limit(lockBatch).Find(&events)

			// 扫表失败
			if result.Error != nil {
				cancelBatch()
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
				cancelBatch()
				logger.Warn("locked outbox events not found", "lock_owner", lockOwner, "locked_rows", lockedRows)
				continue
			}

			// 按 Topic 批量发送，等 Kafka 返回后再回填状态
			sendCtx, cancelSend := context.WithTimeout(batchCtx, sendTimeout)
			sendErrors := sendOutboxEvents(sendCtx, producer, events)
			cancelSend()

			// 逐条按发送结果回填
			var sentCnt, retryCnt, failedCnt, updateFailedCnt, leaseLostCnt int

			for i, e := range events {
				if err := batchCtx.Err(); err != nil {
					logger.Warn("outbox batch interrupted", "error", err, "lock_owner", lockOwner, "remaining", len(events)-i)
					break
				}

				err := sendErrors[i]

				status := model.OutboxEventStatusSent
				updates := map[string]any{
					"lock_owner":    nil,
					"locked_until":  nil,
					"next_retry_at": nil,
				}
				// 发送失败，或超时后无法确认结果，保留原事件供重试
				if err != nil {
					status = model.OutboxEventStatusRetry
					if e.RetryCnt >= 5 {
						status = model.OutboxEventStatusFailed
					} else {
						updates["next_retry_at"] = gorm.Expr("DATE_ADD(NOW(), INTERVAL ? SECOND)", 15)
					}
					updates["retry_cnt"] = gorm.Expr("retry_cnt + ?", 1)

					logger.Error(
						"send outbox event failed",
						"error", err,
						"event_id", e.ID,
						"topic", e.Topic,
						"message_key", e.MessageKey,
						"retry_cnt", e.RetryCnt,
						"next_status", outboxStatusName(status),
					)
				}

				// 只回填还持有有效租约的事件
				updates["status"] = status
				result = batchDB.Model(&model.OutboxEvent{}).
					Where("id = ? AND status = ? AND lock_owner = ? AND locked_until > NOW()", e.ID, model.OutboxEventStatusProcessing, lockOwner).
					Updates(updates)
				if result.Error != nil {
					updateFailedCnt++
					logger.Error(
						"update outbox event status failed",
						"error", result.Error,
						"event_id", e.ID,
						"topic", e.Topic,
						"message_key", e.MessageKey,
						"lock_owner", lockOwner,
						"next_status", outboxStatusName(status),
					)
					continue
				}
				if result.RowsAffected != 1 {
					leaseLostCnt++
					logger.Warn("outbox event lease lost", "event_id", e.ID, "lock_owner", lockOwner)
					continue
				}
				switch status {
				case model.OutboxEventStatusSent:
					sentCnt++
				case model.OutboxEventStatusRetry:
					retryCnt++
				case model.OutboxEventStatusFailed:
					failedCnt++
				default:
				}
			}
			cancelBatch()

			logger.Info(
				"outbox batch processed",
				"lock_owner", lockOwner,
				"locked", lockedRows,
				"loaded", len(events),
				"sent", sentCnt,
				"retry", retryCnt,
				"failed", failedCnt,
				"update_failed", updateFailedCnt,
				"lease_lost", leaseLostCnt,
			)
		}
	}
}

// sendOutboxEvents 按 Topic 并发发送，返回与 events 顺序一致的发送结果。
func sendOutboxEvents(ctx context.Context, producer *kafka.Writer, events []*model.OutboxEvent) []error {
	indexesByTopic := make(map[string][]int)
	for i, e := range events {
		indexesByTopic[e.Topic] = append(indexesByTopic[e.Topic], i)
	}

	sendErrors := make([]error, len(events))
	var wg sync.WaitGroup
	// 每组独立发送，避免单个 Topic 的元数据错误或超时影响其他组。
	// 一轮最多抢占 lockBatch 条，所有组共用本轮发送超时。
	for _, indexes := range indexesByTopic {
		wg.Go(func() {
			messages := make([]kafka.Message, len(indexes))
			for i, index := range indexes {
				e := events[index]
				messages[i] = kafka.Message{
					Topic: e.Topic,
					Key:   []byte(e.MessageKey),
					Value: []byte(e.MessageValue),
				}
			}

			sendErr := producer.WriteMessages(ctx, messages...)
			// Kafka 可能只确认组内一部分消息，按组内下标对应回原事件。
			var writeErrors kafka.WriteErrors
			if errors.As(sendErr, &writeErrors) && len(writeErrors) == len(indexes) {
				for i, index := range indexes {
					sendErrors[index] = writeErrors[i]
				}
				return
			}
			for _, index := range indexes {
				sendErrors[index] = sendErr
			}
		})
	}
	wg.Wait()
	return sendErrors
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
