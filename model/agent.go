package model

import "time"

// Document 原始文档
type Document struct {
	ID        int64
	Title     string     `gorm:"column:title"`      // 文章标题
	Content   string     `gorm:"column:content"`    // 文章内容
	CreatedAt time.Time  `gorm:"column:created_at"` // 创建时间
	UpdatedAt time.Time  `gorm:"column:updated_at"` // 更新时间
	DeletedAt *time.Time `gorm:"column:deleted_at"` // 逻辑删除时间
}

// Chunk 语义片段
type Chunk struct {
	ID        string     `gorm:"column:id"`
	Title     string     `gorm:"column:title"`      // 段落标题
	Content   string     `gorm:"column:content"`    // 段落内容
	Vector    []float64  `gorm:"column:vector"`     // 向量
	CreatedAt time.Time  `gorm:"column:created_at"` // 创建时间
	UpdatedAt time.Time  `gorm:"column:updated_at"` // 更新时间
	DeletedAt *time.Time `gorm:"column:deleted_at"` // 逻辑删除时间
}
