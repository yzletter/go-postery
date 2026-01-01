package model

import "time"

type Order struct {
	ID        int64      `gorm:"primaryKey"`        // 订单 ID
	UserID    int64      `gorm:"column:user_id"`    // 用户 ID
	GiftID    int        `gorm:"column:gift_id"`    // 礼物 ID
	Count     int        `gorm:"column:count"`      // 购买数量
	CreatedAt time.Time  `gorm:"column:created_at"` // 创建时间
	UpdatedAt time.Time  `gorm:"column:updated_at"` // 更新时间
	DeletedAt *time.Time `gorm:"column:deleted_at"` // 逻辑删除时间
}

func (o Order) TableName() string {
	return "orders"
}
