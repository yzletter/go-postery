package model

import (
	"errors"
	"time"
)

// Post 定义数据库模型
type Post struct {
	ID           int64      `gorm:"primaryKey"`           // 帖子 ID
	UserID       int64      `gorm:"column:user_id"`       // 作者 ID
	ViewCount    int        `gorm:"column:view_count"`    // 浏览量
	LikeCount    int        `gorm:"column:like_count"`    // 点赞数
	CommentCount int        `gorm:"column:comment_count"` // 评论数
	Status       int        `gorm:"column:status"`        // 状态 1 正常, 2 封禁
	Title        string     `gorm:"column:title"`         // 标题
	Content      string     `gorm:"column:content"`       // 正文
	CreatedAt    time.Time  `gorm:"column:created_at"`    // 创建时间
	UpdatedAt    time.Time  `gorm:"column:updated_at"`    // 更新时间
	DeletedAt    *time.Time `gorm:"column:deleted_at"`    // 逻辑删除时间
}

// TableName 指定表名
func (p Post) TableName() string {
	return "posts"
}

// PostCntField 用来枚举指定列名
type PostCntField int

const (
	PostViewCount PostCntField = iota + 1
	PostCommentCount
	PostLikeCount
)

func (f PostCntField) Column() (string, error) {
	switch f {
	case PostViewCount:
		return "view_count", nil
	case PostCommentCount:
		return "comment_count", nil
	case PostLikeCount:
		return "like_count", nil
	default:
		return "", errors.New("参数有误")
	}
}

// Redis Key
const (
	KeyPostScore = "post:score"
	KeyPostTime  = "post:time"
)
