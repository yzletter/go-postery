package model

import "time"

type Session struct {
	ID            int64      `gorm:"id,primaryKey"`
	SessionID     int64      `gorm:"session_id"`
	UserID        int64      `gorm:"user_id"`
	TargetID      int64      `gorm:"user_id"`
	TargetType    int        `gorm:"user_id"` // 会话类型 1 表示 私聊 2 表示 群聊
	LastMessageID int64      `gorm:"last_message_id"`
	LastMessage   string     `gorm:"last_message"`
	UnreadCount   int        `gorm:"unread_count"`
	CreatedAt     time.Time  `gorm:"column:created_at"` // 创建时间
	UpdatedAt     time.Time  `gorm:"column:updated_at"` // 更新时间
	DeletedAt     *time.Time `gorm:"column:deleted_at"` // 逻辑删除时间
}

func (s Session) TableName() string {
	return "sessions"
}
