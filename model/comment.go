package model

import "time"

type Comment struct {
	Id         int        `gorm:"primaryKey"`
	PostId     int        `gorm:"column:post_id"`
	ParentId   int        `gorm:"column:parent_id"`
	UserId     int        `gorm:"column:user_id"`
	Content    string     `gorm:"column:content"`
	CreateTime *time.Time `gorm:"column:create_time"`
	DeleteTime *time.Time `gorm:"column:delete_time"`
	UserName   string     `gorm:"-"`
	ViewTime   string     `gorm:"-"`
}
