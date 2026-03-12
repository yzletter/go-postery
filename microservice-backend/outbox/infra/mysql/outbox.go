package infra

import (
	"context"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/yzletter/go-postery/microservice-backend/outbox/model"
	"gorm.io/gorm"
)

func ScanOutbox(ctx context.Context, producer *kafka.Writer) {
	backoff := time.Second  // 退避
	interval := time.Second // 扫表间隔
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("关闭 Scan Outbox 成功 ...")
			return
		case <-ticker.C:
			// 扫表
			var events []*model.Event
			result := globalDB.Model(&model.Event{}).
				Where("status IN (0, 2) AND (next_retry_at IS NULL OR next_retry_at <= NOW())").
				Order("created_at ASC").
				Limit(100).Find(&events)
			if result.Error != nil {
				slog.Error("Scan Outbox Failed", "error", result.Error)

				select {
				case <-ctx.Done():
					slog.Info("关闭 Scan Outbox 成功 ...")
					return
				case <-time.After(backoff):
				}

				if backoff < 4*time.Second {
					backoff *= 2
				}
				continue
			}

			backoff = time.Second

			// Kafka 发消息
			for _, event := range events {
				err := producer.WriteMessages(ctx, kafka.Message{
					Topic: event.Topic,
					Key:   []byte(event.MessageKey),
					Value: []byte(event.MessageValue),
				})
				if err != nil {
					slog.Error("Kafka Write Message Failed", "error", err)
					// 发送失败回填表
					status := 2
					if event.RetryCnt >= 5 { // 已经重试五次了，标记为毒消息
						status = 3
					}
					updates := map[string]any{
						"status":        status,
						"next_retry_at": gorm.Expr("DATE_ADD(NOW(), INTERVAL ? SECOND)", 30),
						"retry_cnt":     gorm.Expr("retry_cnt + ?", 1),
					}

					// 发送失败的更新失败错误可忽略，原消息在数据库中还是未发送的状态
					result = globalDB.Model(&model.Event{}).Where("id = ?", event.ID).Updates(updates)
					if result.Error != nil {
						slog.Error("Update Event Failed", "id", event.ID, "error", result.Error)
					}
					continue
				}

				// 发送成功回填表
				result = globalDB.Model(&model.Event{}).Where("id = ?", event.ID).Update("status", 1)
				if result.Error != nil {
					slog.Error("Update Event Failed", "id", event.ID, "error", result.Error)
				}
			}
		}
	}
}
