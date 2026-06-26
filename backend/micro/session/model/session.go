package model

import "time"

type Session struct {
	ID            int64      `gorm:"column:id;primaryKey"`
	SessionID     int64      `gorm:"column:session_id"`
	UserID        int64      `gorm:"column:user_id"`
	TargetID      int64      `gorm:"column:target_id"`
	TargetType    int        `gorm:"column:target_type"` // 会话类型 1 表示 私聊 2 表示 群聊
	LastMessageID int64      `gorm:"column:last_message_id"`
	LastMessage   string     `gorm:"column:last_message"`
	UnreadCount   int        `gorm:"column:unread_count"`
	CreatedAt     time.Time  `gorm:"column:created_at"` // 创建时间
	UpdatedAt     time.Time  `gorm:"column:updated_at"` // 更新时间
	DeletedAt     *time.Time `gorm:"column:deleted_at"` // 逻辑删除时间
}

func (s Session) TableName() string {
	return "sessions"
}

// UpdateUnread 修改未读
type UpdateUnread struct {
	Updates Updates
	Delta   int
}

type Updates struct {
	LastMessageID   int64     `gorm:"last_message_id"`
	LastMessage     string    `gorm:"last_message"`
	LastMessageTime time.Time `gorm:"updated_at"`
}
