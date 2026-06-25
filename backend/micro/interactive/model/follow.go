package model

import "time"

type Follow struct {
	ID         int64      `gorm:"primaryKey,column:id"`
	FollowerID int64      `gorm:"column:follower_id"`
	FolloweeID int64      `gorm:"column:followee_id"`
	CreatedAt  time.Time  `gorm:"column:created_at"`
	UpdatedAt  time.Time  `gorm:"column:updated_at"`
	DeletedAt  *time.Time `gorm:"column:deleted_at"`
}

// TableName 指定表名
func (f Follow) TableName() string {
	return "follows"
}
