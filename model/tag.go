package model

import "time"

type Tag struct {
	ID        int64      `gorm:"primaryKey"` // ID 雪花算法
	Name      string     `gorm:"column:name"`
	Slug      string     `gorm:"column:slug"` // 标签唯一标识
	CreatedAt time.Time  `gorm:"column:created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at"`
}

type PostTag struct {
	ID        int64      `gorm:"column:id"`
	PostID    int64      `gorm:"column:post_id"`
	TagID     int64      `gorm:"column:tag_id"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at"`
}

func (pt PostTag) TableName() string {
	return "post_tag"
}

func (t Tag) TableName() string {
	return "tags"
}
