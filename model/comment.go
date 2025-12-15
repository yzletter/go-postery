package model

import "time"

type Comment struct {
	ID        int64      `gorm:"primaryKey"`
	PostID    int64      `gorm:"column:post_id"`
	ParentID  int64      `gorm:"column:parent_id"`
	ReplyID   int64      `gorm:"column:reply_id"`
	UserID    int64      `gorm:"column:user_id"`
	Content   string     `gorm:"column:content"`
	CreatedAt time.Time  `gorm:"column:created_at"` // 创建时间
	UpdatedAt time.Time  `gorm:"column:updated_at"` // 更新时间
	DeletedAt *time.Time `gorm:"column:deleted_at"` // 逻辑删除时间
}

// TableName 指定表名
func (c Comment) TableName() string {
	return "comments"
}
