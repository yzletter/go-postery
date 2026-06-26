package lottery

import (
	lottery_grpc "github.com/yzletter/go-postery/api/proto/lottery/v1"
	user_grpc "github.com/yzletter/go-postery/api/proto/user/v1"
	userdto "github.com/yzletter/go-postery/backend/bff/dto/user"
)

type GiftDTO struct {
	ID          int64  `json:"id,string"`
	Name        string `json:"name"`
	Avatar      string `json:"avatar"`
	Description string `json:"description"`
	Prize       int    `json:"prize"`
}

func ToGiftDTO(gift *lottery_grpc.Gift) GiftDTO {
	if gift == nil {
		return GiftDTO{}
	}
	return GiftDTO{
		ID:          gift.ID,
		Name:        gift.Name,
		Avatar:      gift.Avatar,
		Description: gift.Description,
		Prize:       int(gift.Prize),
	}
}

type LotteryResultDTO struct {
	GiftDTO
	TempOrderID       int64  `json:"temp_order_id,string,omitempty"`
	Success           bool   `json:"success"`
	ResultDescription string `json:"result_description,omitempty"`
	UserID            int64  `json:"user_id,string,omitempty"`
}

func ToLotteryResultDTO(result *lottery_grpc.LotteryResponse) LotteryResultDTO {
	if result == nil {
		return LotteryResultDTO{}
	}
	gift := ToGiftDTO(result.Gift)
	if gift.ID == 0 && gift.Name == "" && result.Description != "" {
		gift = GiftDTO{
			ID:          0,
			Name:        lotteryResultName(result.Description),
			Description: result.Description,
		}
	}
	return LotteryResultDTO{
		GiftDTO:           gift,
		TempOrderID:       result.TempOrderID,
		Success:           result.Success,
		ResultDescription: result.Description,
		UserID:            result.UserID,
	}
}

func lotteryResultName(description string) string {
	switch description {
	case "奖品已抽完":
		return "奖品已抽完"
	case "很遗憾，未抽中奖品，谢谢参与":
		return "谢谢参与"
	default:
		return description
	}
}

type OrderDTO struct {
	ID        int64            `json:"id,string"` // 订单 ID
	User      userdto.BriefDTO `json:"user"`
	Gift      GiftDTO          `json:"gift"`
	Count     int              `json:"count"`      // 购买数量
	CreatedAt string           `json:"created_at"` // 创建时间
	Status    int32            `json:"status"`     // 订单状态 0 待支付，1 已支付，2 已放弃，3 已超时
	PaidAt    string           `json:"paid_at"`    // 支付时间
	ExpireAt  string           `json:"expire_at"`  // 过期时间
}

func ToOrderDTO(order *lottery_grpc.Order, user *user_grpc.Profile) OrderDTO {
	return OrderDTO{
		ID:        order.OrderID,
		User:      userdto.ToBriefDTO(user),
		Gift:      ToGiftDTO(order.Gift),
		Count:     int(order.Count),
		CreatedAt: order.CreatedAt,
		Status:    order.Status,
		PaidAt:    order.PaidAt,
		ExpireAt:  order.ExpireAt,
	}
}
