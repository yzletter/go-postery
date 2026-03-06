package dto

import (
	"time"

	lottery_grpc "github.com/yzletter/go-postery/api/proto/lottery/v1"
	model2 "github.com/yzletter/go-postery/microservice-backend/lottery/model"
)

// ToGift model 转 lottery_grpc.Gift
func ToGift(gift *model2.Gift) *lottery_grpc.Gift {
	return &lottery_grpc.Gift{
		ID:          gift.ID,
		Name:        gift.Name,
		Avatar:      gift.Avatar,
		Description: gift.Description,
		Prize:       int64(gift.Prize),
	}
}

// ToOrder model 转 lottery_grpc.Order
func ToOrder(order *model2.Order, gift *model2.Gift) *lottery_grpc.Order {
	return &lottery_grpc.Order{
		OrderID:   order.ID,
		UserID:    order.UserID,
		Gift:      ToGift(gift),
		Count:     int64(order.Count),
		CreatedAt: order.CreatedAt.Format(time.RFC3339),
	}
}
