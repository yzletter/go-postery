package dto

import (
	"time"

	lottery_grpc "github.com/yzletter/go-postery/api/proto/lottery/v1"
	"github.com/yzletter/go-postery/backend/micro/lottery/model"
)

// ToGift model 转 lottery_grpc.Gift
func ToGift(gift *model.Gift) *lottery_grpc.Gift {
	if gift == nil {
		return &lottery_grpc.Gift{}
	}
	return &lottery_grpc.Gift{
		ID:          gift.ID,
		Name:        gift.Name,
		Avatar:      gift.Avatar,
		Description: gift.Description,
		Prize:       int64(gift.Prize),
	}
}

func ToLotteryResponse(result *LotteryResult) *lottery_grpc.LotteryResponse {
	if result == nil {
		return &lottery_grpc.LotteryResponse{}
	}
	return &lottery_grpc.LotteryResponse{
		Gift:        ToGift(result.Gift),
		TempOrderID: result.OrderID,
		Success:     result.Success,
		Description: result.Description,
		UserID:      result.UserID,
	}
}

// ToOrder model 转 lottery_grpc.Order
func ToOrder(order *model.Order, gift *model.Gift) *lottery_grpc.Order {
	if order == nil {
		return &lottery_grpc.Order{
			Gift: ToGift(gift),
		}
	}

	return &lottery_grpc.Order{
		OrderID:   order.ID,
		UserID:    order.UserID,
		Gift:      ToGift(gift),
		Count:     int64(order.Count),
		CreatedAt: formatTime(order.CreatedAt),
		Status:    int32(order.Status),
		PaidAt:    formatTimePtr(order.PaidAt),
		ExpireAt:  formatTime(order.ExpireAt),
	}
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

func formatTimePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return formatTime(*t)
}
