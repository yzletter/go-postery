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
	ID        string     `gorm:"column:id" json:"id"`
	Title     string     `gorm:"-" json:"title"`                // 段落标题 暂时搁置
	Content   string     `gorm:"column:content" json:"content"` // 段落内容
	BatchID   int64      `gorm:"column:batch_id"`
	CreatedAt time.Time  `gorm:"column:created_at" json:"createdAt"` // 创建时间
	UpdatedAt time.Time  `gorm:"column:updated_at" json:"updatedAt"` // 更新时间
	DeletedAt *time.Time `gorm:"column:deleted_at" json:"deletedAt"` // 逻辑删除时间
}

func (c Chunk) TableName() string {
	return "chunks"
}
