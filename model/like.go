package model

import "time"

type UserLike struct {
	Id         int        `gorm:"primaryKey"`
	UserId     int        `gorm:"column:user_id"`
	PostId     int        `gorm:"column:post_id"`
	CreateTime *time.Time `gorm:"column:create_time"`
	UpdateTime *time.Time `gorm:"column:update_time"`
	DeleteTime *time.Time `gorm:"column:delete_time"`
}

func (u UserLike) TableName() string {
	return "user_like"
}
