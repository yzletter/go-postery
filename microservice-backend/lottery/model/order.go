package model

import "time"

type Order struct {
	ID                      int64      `gorm:"primaryKey"`                        // 订单 ID
	UserID                  int64      `gorm:"column:user_id"`                    // 用户 ID
	GiftID                  int64      `gorm:"column:gift_id"`                    // 礼物 ID
	Count                   int        `gorm:"column:count;default:1"`            // 购买数量
	ExpireAt                time.Time  `gorm:"column:expire_at"`                  // 过期时间
	PaidAt                  *time.Time `gorm:"column:paid_at"`                    // 支付时间
	Status                  int        `gorm:"column:status"`                     // 订单状态 0 待支付，1 已支付，2 已放弃，3 已超时
	StockRollbackStatus     int        `gorm:"column:stock_rollback_status"`      // 回补状态 0 待回补，1 已回补，2 回补失败
	StockRollbackRetryCount int        `gorm:"column:stock_rollback_retry_count"` // 已尝试回补次数
	NextRollbackAt          time.Time  `gorm:"column:next_rollback_at"`           // 下次回补时间
	CreatedAt               time.Time  `gorm:"column:created_at"`                 // 创建时间
	UpdatedAt               time.Time  `gorm:"column:updated_at"`                 // 更新时间
	DeletedAt               *time.Time `gorm:"column:deleted_at"`                 // 逻辑删除时间
}

const (
	OrderStatusPending = iota
	OrderStatusPaid
	OrderStatusCancelled
	OrderStatusExpired
)

const (
	StockRollbackStatusPending = iota
	StockRollbackStatusDone
	StockRollbackStatusFailed
)

func (o Order) TableName() string {
	return "orders"
}
