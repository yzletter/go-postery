package model

import "time"

type Post struct {
	Id         int        `gorm:"primaryKey"`
	UserId     int        `gorm:"column:user_id"`
	Title      string     `gorm:"column:title"`
	Content    string     `gorm:"column:content"`
	CreateTime *time.Time `gorm:"column:create_time"`
	DeleteTime *time.Time `gorm:"column:delete_time"`
	ViewCount  int        `gorm:"column:view_count"`
}
