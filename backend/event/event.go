package event

import "time"

// Event 待投递消息事件
type Event struct {
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

func (e Event) TableName() string {
	return "events"
}

const (
	StatusEventPending    = iota // 待发送
	StatusEventProcessing        // 发送中
	StatusEventSent              // 已发送
	StatusEventRetry             // 需重试
	StatusEventFailed            // 失败
)

type RegisterSessionEventPayload struct {
	UserID int64 `json:"user_id,string"`
}

type InitUserScoreEventPayload struct {
	UserID int64 `json:"user_id,string"`
}

type ChunkDocumentEventPayload struct {
	ID int64 `json:"id,string"`
}

type UpsertQdrantEventPayload struct {
	BatchID int64 `json:"batch_id,string"`
}
