package model

import "time"

type Comment struct {
	ID        int64      `gorm:"primaryKey"`        // 评论 ID
	PostID    int64      `gorm:"column:post_id"`    // 帖子 ID
	ParentID  int64      `gorm:"column:parent_id"`  // 父评论 ID, 可为空
	ReplyID   int64      `gorm:"column:reply_id"`   // 回复的评论 ID, 可为空
	UserID    int64      `gorm:"column:user_id"`    // 评论者 ID
	Content   string     `gorm:"column:content"`    // 评论内容
	CreatedAt time.Time  `gorm:"column:created_at"` // 创建时间
	UpdatedAt time.Time  `gorm:"column:updated_at"` // 更新时间
	DeletedAt *time.Time `gorm:"column:deleted_at"` // 逻辑删除时间
}

// TableName 指定表名
func (c Comment) TableName() string {
	return "comments"
}
