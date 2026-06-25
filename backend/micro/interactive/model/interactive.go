package model

import "time"

// PostInteractive 帖子互动表
type PostInteractive struct {
	ID          int64      `gorm:"primary_key"`
	PostID      int64      `gorm:"column:post_id"`
	ReadCnt     int64      `gorm:"column:read_count"`
	LikeCnt     int64      `gorm:"column:like_count"`
	CommentCnt  int64      `gorm:"column:comment_count"`
	CalculateAt *time.Time `gorm:"column:calculate_at"` // 上次扫表计算 Inter 的时间
	CreatedAt   time.Time  `gorm:"column:created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at"`
	DeletedAt   *time.Time `gorm:"column:deleted_at"`
}

func (p PostInteractive) TableName() string {
	return "post_interactive"
}

// UserInteractive 用户互动表
type UserInteractive struct {
	ID          int64      `gorm:"primary_key"`
	UserID      int64      `gorm:"column:user_id"`
	FollowCnt   int64      `gorm:"column:follow_count"`
	CalculateAt *time.Time `gorm:"column:calculate_at"` // 上次扫表计算 Inter 的时间
	CreatedAt   time.Time  `gorm:"column:created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at"`
	DeletedAt   *time.Time `gorm:"column:deleted_at"`
}

func (u UserInteractive) TableName() string {
	return "user_interactive"
}
