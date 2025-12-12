package model

import "time"

type Post struct {
	Id           int `gorm:"primaryKey"`
	UserId       int `gorm:"column:user_id"`       // 作者 id
	ViewCount    int `gorm:"column:view_count"`    // 浏览量
	LikeCount    int `gorm:"column:like_count"`    // 点赞数
	CommentCount int `gorm:"column:comment_count"` // 评论数

	Title   string `gorm:"column:title"`
	Content string `gorm:"column:content"`

	CreateTime *time.Time `gorm:"column:create_time"`
	UpdateTime *time.Time `gorm:"column:update_time"`
	DeleteTime *time.Time `gorm:"column:delete_time"`
}
