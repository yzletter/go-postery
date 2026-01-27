package model

import "time"

type Like struct {
	ID        int64      `gorm:"primaryKey"`
	UserID    int64      `gorm:"column:user_id"`
	PostID    int64      `gorm:"column:post_id"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at"`
}

func (l Like) TableName() string {
	return "likes"
}
