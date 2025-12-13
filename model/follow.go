package model

import "time"

type Follow struct {
	Id         int        `gorm:"primaryKey"`
	FollowerId int        `gorm:"follower_id"`
	FolloweeId int        `gorm:"followee_id"`
	CreateTime *time.Time `gorm:"column:create_time"`
	UpdateTime *time.Time `gorm:"column:update_time"`
	DeleteTime *time.Time `gorm:"column:delete_time"`
}
