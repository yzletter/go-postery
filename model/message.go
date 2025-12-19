package model

import "time"

type Message struct {
	ID          int64      `gorm:"id,primaryKey"`
	MessageFrom int64      `gorm:"message_from"`
	MessageTo   int64      `gorm:"message_to"`
	Content     string     `gorm:"content"`
	CreatedAt   time.Time  `gorm:"column:created_at"` // 创建时间
	UpdatedAt   time.Time  `gorm:"column:updated_at"` // 更新时间
	DeletedAt   *time.Time `gorm:"column:deleted_at"` // 逻辑删除时间
}
