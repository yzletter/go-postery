package model

import "time"

type Tag struct {
	Id         int        `gorm:"primaryKey"` // ID 雪花算法
	Name       string     `gorm:"column:name"`
	Slug       string     `gorm:"column:slug"` // 标签唯一标识
	CreateTime *time.Time `gorm:"column:create_time"`
	DeleteTime *time.Time `gorm:"column:delete_time"`
}

type PostTag struct {
	Id     int `gorm:"column:id"`
	PostId int `gorm:"column:post_id"`
	TagId  int `gorm:"column:tag_id"`
}

func (pt PostTag) TableName() string {
	return "post_tag"
}
