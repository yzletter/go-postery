package dto

import (
	"github.com/yzletter/go-postery/backend/micro/lottery/model"
)

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
