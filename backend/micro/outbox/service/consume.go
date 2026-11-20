package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/yzletter/go-postery/backend/conf"
	infraKafka "github.com/yzletter/go-postery/backend/infra/mq/kafka"
	"github.com/yzletter/go-postery/backend/micro/outbox/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

const (
	maxConsumeAttempts = 3                // 首次处理加两次重试
	consumeTimeout     = 10 * time.Second // 单次业务处理的上下文超时
	batchWait          = 5 * time.Second  // 从第一条消息开始等待凑批
)

// ErrInvalidMessage 消息格式或字段无效，直接写死信表。
var ErrInvalidMessage = errors.New("invalid kafka message")

// Consumer 管理一个消费组的分区消费和死信回放。
type Consumer struct {
	brokers []string
	groupID string
	topics  []string
	db      *gorm.DB
	done    chan struct{}
}

// NewConsumer 创建负责业务消费和死信回放的 Consumer。
func NewConsumer(config conf.KafkaConfig, db *gorm.DB, topic string, group string, topics ...string) *Consumer {
	return &Consumer{
		brokers: []string{config.Addr},
		groupID: group,
		topics:  append([]string{topic}, topics...),
		db:      db,
		done:    make(chan struct{}),
	}
}

// Wait 等待消费和回放协程退出，在关闭数据库前调用。
func (consumer *Consumer) Wait() {
	timer := time.NewTimer(10 * time.Second)
	defer timer.Stop()
	select {
	case <-consumer.done:
	case <-timer.C:
		slog.Warn("wait for kafka consumer timed out", "group", consumer.groupID)
	}
}

// DecodeMessage 反序列化消息，将解析失败标记为无效消息。
func DecodeMessage(msg kafka.Message, payload any) error {
	if err := sonic.Unmarshal(msg.Value, payload); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidMessage, err)
	}
	return nil
}

// Run 在当前服务执行正常消费和人工回放，两者使用同一个业务处理函数。
func (consumer *Consumer) Run(ctx context.Context, handler func(context.Context, ...kafka.Message) error, batchSizes map[string]int) {
	defer close(consumer.done)
	group, err := infraKafka.InitConsumer(consumer.brokers, consumer.groupID, consumer.topics...)
	if err != nil {
		slog.Error("create kafka consumer group failed", "group", consumer.groupID, "error", err)
		return
	}

	// 独立扫死信表，消费组重新分配分区时仍可继续回放
	replayCtx, cancelReplay := context.WithCancel(ctx)
	replayDone := make(chan struct{})
	go func() {
		defer close(replayDone)
		consumer.runReplay(replayCtx, handler)
	}()
	defer func() {
		cancelReplay()
		_ = group.Close()
		<-replayDone
	}()

	// 每次分配创建对应的分区协程。分配结束后，旧协程会退出。
	for ctx.Err() == nil {
		generation, err := group.Next(ctx)
		if err != nil {
			if ctx.Err() != nil || errors.Is(err, kafka.ErrGroupClosed) {
				return
			}
			slog.Warn("join kafka consumer group failed", "group", consumer.groupID, "error", err)
			wait(ctx, time.Second)
			continue
		}
		for topic, partitions := range generation.Assignments {
			for _, partition := range partitions {
				generation.Start(func(generationCtx context.Context) {
					consumer.consumePartition(generationCtx, generation, topic, partition, max(batchSizes[topic], 1), handler)
				})
			}
		}
	}
}

// consumePartition 按顺序处理一个分区，提交成功后再拉取下一条或下一批。
func (consumer *Consumer) consumePartition(ctx context.Context, generation *kafka.Generation, topic string, partition kafka.PartitionAssignment, batchSize int, handler func(context.Context, ...kafka.Message) error) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   consumer.brokers,
		Topic:     topic,
		Partition: partition.ID,
	})
	defer reader.Close()
	logger := slog.With("group", consumer.groupID, "topic", topic, "partition", partition.ID)

	// 从 Kafka 分配的位点开始。首次消费时使用配置的初始位点。
	if err := reader.SetOffset(partition.Offset); err != nil {
		logger.Error("set kafka partition offset failed", "error", err)
		return
	}

	messages := make([]kafka.Message, 0, batchSize)
	for ctx.Err() == nil {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			logger.Warn("fetch kafka message failed", "error", err)
			wait(ctx, time.Second)
			continue
		}
		messages = append(messages[:0], msg)

		// 从批次第一条开始计时，避免持续的小流量让批次一直等不到满。
		if batchSize > 1 {
			batchCtx, cancel := context.WithTimeout(ctx, batchWait)
			for len(messages) < batchSize {
				msg, err := reader.FetchMessage(batchCtx)
				if err != nil {
					break
				}
				messages = append(messages, msg)
			}
			cancel()
		}

		// 死信写入失败时退出本轮分配，避免继续提交更靠后的位点。
		if err := consumer.handleMessages(ctx, handler, messages); err != nil {
			if ctx.Err() == nil {
				logger.Error("handle kafka messages failed without committing", "offset", messages[0].Offset, "error", err)
				wait(ctx, time.Second)
			}
			return // 下次分配会从 Kafka 已提交的位点重新读取
		}

		// 整批处理成功或死信已落库，提交下一条消息的位点。
		if ctx.Err() != nil {
			return
		}
		offset := messages[len(messages)-1].Offset + 1
		if err := generation.CommitOffsets(map[string]map[int]int64{topic: {partition.ID: offset}}); err != nil {
			logger.Error("commit kafka offset failed", "next_offset", offset, "error", err)
			wait(ctx, time.Second)
			return // 提交失败时重新获取分配，已处理的消息可能被重放
		}
	}
}

// handleMessages 处理当前批次。处理成功或确认死信已落库后，调用方才可以提交位点。
func (consumer *Consumer) handleMessages(ctx context.Context, handler func(context.Context, ...kafka.Message) error, messages []kafka.Message) error {
	if len(messages) > 1 {
		if err := callHandler(ctx, handler, messages...); err == nil {
			return nil
		}
		// 批次失败后逐条处理，隔离坏消息。业务幂等负责处理可能的重放。
	}
	for _, msg := range messages {
		err := retryMessage(ctx, handler, msg)
		if err == nil {
			continue
		}
		// 消费组分配结束时直接退出，留给下一轮重新读取。
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if saveErr := consumer.saveFailedMessage(ctx, msg, err); saveErr != nil {
			return fmt.Errorf("save failed kafka message at offset %d: %w", msg.Offset, saveErr)
		}
	}
	return nil
}

// retryMessage 原地重试当前消息，两次重试分别等待 1 秒和 2 秒。
func retryMessage(ctx context.Context, handler func(context.Context, ...kafka.Message) error, msg kafka.Message) error {
	backoff := time.Second
	for attempt := 1; ; attempt++ {
		err := callHandler(ctx, handler, msg)
		if err == nil || errors.Is(err, ErrInvalidMessage) || attempt == maxConsumeAttempts {
			return err
		}
		if !wait(ctx, backoff) {
			return ctx.Err()
		}
		backoff *= 2
	}
}

// callHandler 为单次业务处理设置独立超时，继承消费或回放的取消信号。
func callHandler(ctx context.Context, handler func(context.Context, ...kafka.Message) error, messages ...kafka.Message) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	attemptCtx, cancel := context.WithTimeout(ctx, consumeTimeout)
	defer cancel()
	return handler(attemptCtx, messages...)
}

func wait(ctx context.Context, delay time.Duration) bool {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

// saveFailedMessage 将处理失败的消息写入死信表，交由人工排查
func (consumer *Consumer) saveFailedMessage(ctx context.Context, msg kafka.Message, cause error) error {
	headers, err := sonic.MarshalString(msg.Headers)
	if err != nil {
		return err
	}

	saveCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	event := model.FailedEvent{
		ConsumerGroup: consumer.groupID,
		Topic:         msg.Topic,
		Partition:     msg.Partition,
		Offset:        msg.Offset,
		MessageKey:    msg.Key,
		MessageValue:  msg.Value,
		Headers:       headers,
		LastError:     strings.ToValidUTF8(cause.Error(), "�"),
		Status:        model.FailedEventStatusManualRequired,
	}
	// 通过消费组、Topic、分区和位点去重，保留原记录的回放状态。
	// 消息原文和 Header 只写死信表，避免记录到 SQL 日志。
	result := consumer.db.WithContext(saveCtx).
		Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)}).
		Clauses(clause.OnConflict{DoNothing: true}).Create(&event)
	if result.Error != nil {
		return result.Error
	}
	slog.Warn("kafka message saved for manual replay", "group", event.ConsumerGroup, "topic", msg.Topic,
		"partition", msg.Partition, "offset", msg.Offset, "error", cause)
	return nil
}

// runReplay 定时处理人工发起的回放，每轮最多领取一条
func (consumer *Consumer) runReplay(ctx context.Context, handler func(context.Context, ...kafka.Message) error) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			replayCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			err := consumer.replayOne(replayCtx, handler)
			cancel()
			if err != nil && ctx.Err() == nil {
				slog.Error("replay kafka message failed", "group", consumer.groupID, "error", err)
			}
		}
	}
}

// replayOne 领取当前消费组的一条死信并回放一次
func (consumer *Consumer) replayOne(ctx context.Context, handler func(context.Context, ...kafka.Message) error) error {
	db := consumer.db.WithContext(ctx).Model(&model.FailedEvent{})

	// 回放进程中断时结果未知，租约到期后恢复为待人工确认。
	if err := db.Session(&gorm.Session{}).
		Where("consumer_group = ? AND status = ? AND locked_until <= NOW()", consumer.groupID, model.FailedEventStatusReplaying).
		Updates(map[string]any{
			"status":       model.FailedEventStatusManualRequired,
			"last_error":   "replay interrupted, result unknown",
			"lock_owner":   nil,
			"locked_until": nil,
		}).Error; err != nil {
		return err
	}

	// 条件更新抢占待回放记录，防止多个实例同时执行同一条回放
	lockOwner := uuid.NewString()
	result := db.Session(&gorm.Session{}).
		Where("consumer_group = ? AND status = ?", consumer.groupID, model.FailedEventStatusReplayPending).
		Order("id ASC").Limit(1).
		Updates(map[string]any{
			"status":       model.FailedEventStatusReplaying,
			"replay_count": gorm.Expr("replay_count + 1"),
			"lock_owner":   lockOwner,
			"locked_until": gorm.Expr("DATE_ADD(NOW(), INTERVAL 60 SECOND)"),
		})
	if result.Error != nil || result.RowsAffected == 0 {
		return result.Error
	}

	var event model.FailedEvent
	if err := db.Session(&gorm.Session{}).
		Where("consumer_group = ? AND status = ? AND lock_owner = ? AND locked_until > NOW()",
			consumer.groupID, model.FailedEventStatusReplaying, lockOwner).
		First(&event).Error; err != nil {
		return err
	}

	msg := kafka.Message{
		Topic: event.Topic, Partition: event.Partition, Offset: event.Offset,
		Key: event.MessageKey, Value: event.MessageValue,
	}
	err := sonic.UnmarshalString(event.Headers, &msg.Headers)
	if err == nil {
		// 使用原消息处理逻辑，保留消息里的事件 ID 和原消费组。
		err = callHandler(ctx, handler, msg)
	}
	status, lastError := model.FailedEventStatusResolved, ""
	if err != nil {
		status = model.FailedEventStatusManualRequired
		lastError = strings.ToValidUTF8(err.Error(), "�")
	}

	// 按持有者和租约回填原记录。回放失败后等待下一次人工触发
	result = db.Session(&gorm.Session{}).
		Where("id = ? AND status = ? AND lock_owner = ? AND locked_until > NOW()", event.ID, model.FailedEventStatusReplaying, lockOwner).
		Updates(map[string]any{
			"status":       status,
			"last_error":   lastError,
			"lock_owner":   nil,
			"locked_until": nil,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return errors.New("kafka replay lease lost")
	}
	if err != nil {
		slog.Warn("kafka replay requires manual attention", "group", consumer.groupID, "id", event.ID, "error", err)
	} else {
		slog.Info("kafka replay finished", "group", consumer.groupID, "id", event.ID, "status", status)
	}
	return nil
}
