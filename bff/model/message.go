package model

import "time"

type Message struct {
	ID          int64      `json:"id,string" gorm:"id,primaryKey"`
	SessionID   int64      `json:"session_id,string" gorm:"session_id"`
	SessionType int        `json:"session_type" gorm:"session_type"`
	MessageFrom int64      `json:"message_from,string" gorm:"message_from"`
	MessageTo   int64      `json:"message_to,string" gorm:"message_to"`
	Content     string     `json:"content" gorm:"content"`
	CreatedAt   time.Time  `json:"created_at" gorm:"column:created_at"` // 创建时间
	UpdatedAt   time.Time  `json:"updated_at" gorm:"column:updated_at"` // 更新时间
	DeletedAt   *time.Time `json:"deleted_at" gorm:"column:deleted_at"` // 逻辑删除时间
}
