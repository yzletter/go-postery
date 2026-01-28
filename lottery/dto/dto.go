package dto

import (
	"time"

	lottery_grpc "github.com/yzletter/go-postery/api/proto/lottery/v1"
	"github.com/yzletter/go-postery/lottery/model"
)

// ToGift model 转 lottery_grpc.Gift
func ToGift(gift *model.Gift) *lottery_grpc.Gift {
	return &lottery_grpc.Gift{
		ID:          gift.ID,
		Name:        gift.Name,
		Avatar:      gift.Avatar,
		Description: gift.Description,
		Prize:       int64(gift.Prize),
	}
}

// ToOrder model 转 lottery_grpc.Order
func ToOrder(order *model.Order, gift *model.Gift) *lottery_grpc.Order {
	return &lottery_grpc.Order{
		OrderID:   order.ID,
		UserID:    order.UserID,
		Gift:      ToGift(gift),
		Count:     int64(order.Count),
		CreatedAt: order.CreatedAt.Format(time.RFC3339),
	}
}
