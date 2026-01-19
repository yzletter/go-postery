package model

import "time"

type Event struct {
	ID           int64      `gorm:"column:id;primaryKey"`
	Status       int        `gorm:"column:status"`
	RetryCnt     int        `gorm:"column:retry_cnt"`
	Topic        string     `gorm:"column:topic"`
	MessageKey   string     `gorm:"column:message_key"`
	MessageValue string     `gorm:"column:message_value"`
	NextRetryAt  *time.Time `gorm:"column:next_retry_at"`
	CreatedAt    time.Time  `gorm:"column:created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at"`
}

func (e Event) TableName() string {
	return "events"
}

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
	IDs     []string
	Vectors [][]float64
}
