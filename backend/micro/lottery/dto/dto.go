package dto

import (
	"github.com/yzletter/go-postery/backend/api/proto/lottery/v1"
	"github.com/yzletter/go-postery/backend/micro/lottery/model"
	"github.com/yzletter/go-postery/backend/utils"
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

// ToLotteryResponse LotteryResult 转 lottery_grpc.LotteryResponse
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
		CreatedAt: utils.FormatTime(order.CreatedAt),
		Status:    int32(order.Status),
		PaidAt:    utils.FormatTimePtr(order.PaidAt),
		ExpireAt:  utils.FormatTime(order.ExpireAt),
	}
}
