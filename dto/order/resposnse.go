package order

import (
	"time"

	"github.com/yzletter/go-postery/model"
)

type DTO struct {
	ID        int64  `json:"id,string"`  // 订单 ID
	UserID    int64  `json:"user_id"`    // 用户 ID
	GiftID    int64  `json:"gift_id"`    // 礼物 ID
	Count     int    `json:"count"`      // 购买数量
	CreatedAt string `json:"created_at"` // 创建时间
}

func ToDTO(order *model.Order) DTO {
	return DTO{
		ID:        order.ID,
		UserID:    order.UserID,
		GiftID:    order.GiftID,
		Count:     order.Count,
		CreatedAt: order.CreatedAt.Format(time.RFC3339),
	}
}
