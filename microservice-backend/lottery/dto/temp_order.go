package dto

import (
	"github.com/yzletter/go-postery/microservice-backend/lottery/model"
)

type Order struct {
	ID     int64 `json:"id,string"`
	UserID int64 `json:"user_id,string"`
	GiftID int64 `json:"gift_id,string"`
}

type LotteryResult struct {
	Success     bool   `json:"success"`
	Description string `json:"description"`
	OrderID     int64  `json:"order_id,string"`
	UserID      int64  `json:"user_id,string"`
	Gift        *model.Gift
}

const (
	DescriptionNoGifts        = "奖品已抽完"
	DescriptionLotterySuccess = "抽奖成功"
	DescriptionLotteryNothing = "很遗憾，未抽中奖品，谢谢参与"
	DescriptionTempOrderToPay = "当前已有订单"
)
