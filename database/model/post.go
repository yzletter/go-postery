package model

import "time"

type Post struct {
	Id         int        `gorm:"primaryKey"`
	UserId     int        `gorm:"column:user_id"`
	Tittle     string     `gorm:"column:tittle"`
	Content    string     `gorm:"column:content"`
	CreateTime *time.Time `gorm:"column:create_time"`
	DeleteTime *time.Time `gorm:"column:delete_time"`
	ViewTime   string     `gorm:"-"` // 前端用于展示的时间
}
