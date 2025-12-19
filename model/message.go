package model

import "time"

//id BIGINT NOT NULL COMMENT 'ID',
//
//message_from BIGINT NOT NULL COMMENT '发送方',
//message_to BIGINT NOT NULL COMMENT '接收方',
//
//content TEXT NOT NULL COMMENT '消息内容',
//created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
//updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
//deleted_at DATETIME DEFAULT NULL COMMENT '逻辑删除时间',

type Message struct {
	ID          int64      `gorm:"id,primaryKey"`
	MessageFrom int64      `gorm:"message_from"`
	MessageTo   int64      `gorm:"message_to"`
	Content     string     `gorm:"content"`
	CreatedAt   time.Time  `gorm:"column:created_at"` // 创建时间
	UpdatedAt   time.Time  `gorm:"column:updated_at"` // 更新时间
	DeletedAt   *time.Time `gorm:"column:deleted_at"` // 逻辑删除时间
}
