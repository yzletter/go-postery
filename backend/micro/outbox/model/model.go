package model

import "time"

// OutboxEvent 发送状态
const (
	OutboxEventStatusPending    = iota // 待发送
	OutboxEventStatusProcessing        // 发送中
	OutboxEventStatusSent              // 已发送
	OutboxEventStatusRetry             // 需重试
	OutboxEventStatusFailed            // 失败
)

// OutboxEvent 待投递 Outbox 消息事件
type OutboxEvent struct {
	ID           int64      `gorm:"column:id;primaryKey"` // 消息 ID
	Status       int        `gorm:"column:status"`        // 消息发送状态 0 待发送, 1 发送中, 2 已发送, 3 需重试, 4 失败
	RetryCnt     int        `gorm:"column:retry_cnt"`     // 已重试次数
	Topic        string     `gorm:"column:topic"`         // 消息 Topic
	MessageKey   string     `gorm:"column:message_key"`   // 消息 Key
	MessageValue string     `gorm:"column:message_value"` // 消息内容
	NextRetryAt  *time.Time `gorm:"column:next_retry_at"` // 下一次重试时间
	LockOwner    string     `gorm:"column:lock_owner"`    // 当前处理者
	LockedUntil  *time.Time `gorm:"column:locked_until"`  // 锁过期时间
	CreatedAt    time.Time  `gorm:"column:created_at"`    // 创建时间
	UpdatedAt    time.Time  `gorm:"column:updated_at"`    // 更新时间
}

func (e OutboxEvent) TableName() string {
	return "events"
}

// ProcessedEvent 消费成功同时写幂等表
type ProcessedEvent struct {
	ID        int64     `gorm:"column:id;primaryKey"`
	Consumer  string    `gorm:"column:consumer"`
	EventID   int64     `gorm:"column:event_id"`
	Topic     string    `gorm:"column:topic"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (e ProcessedEvent) TableName() string {
	return "processed_events"
}

// FailedEvent 回放状态
const (
	FailedEventStatusManualRequired = "manual_required" // 等待人工排查
	FailedEventStatusReplayPending  = "replay_pending"  // 已人工发起回放，等待领取
	FailedEventStatusReplaying      = "replaying"       // 已领取，回放中
	FailedEventStatusResolved       = "resolved"        // 业务处理成功
)

// FailedEvent 保存原始 Kafka 消息和最近一次处理错误，供人工回放。
type FailedEvent struct {
	ID            int64  `gorm:"column:id;primaryKey;autoIncrement"`
	ConsumerGroup string `gorm:"column:consumer_group"`
	Topic         string `gorm:"column:topic"`
	Partition     int    `gorm:"column:partition_id"`
	Offset        int64  `gorm:"column:message_offset"`
	MessageKey    []byte `gorm:"column:message_key"`
	MessageValue  []byte `gorm:"column:message_value"`
	Headers       string `gorm:"column:message_headers"`
	LastError     string `gorm:"column:last_error"`
	Status        string `gorm:"column:status"`
}

func (e FailedEvent) TableName() string {
	return "consumer_failed_events"
}
