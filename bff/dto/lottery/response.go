package lottery

import (
	lottery_grpc "github.com/yzletter/go-postery/api/proto/lottery/v1"
	user_grpc "github.com/yzletter/go-postery/api/proto/user/v1"
	userdto "github.com/yzletter/go-postery/bff/dto/user"
)

type GiftDTO struct {
	ID          int64  `json:"id,string"`
	Name        string `json:"name"`
	Avatar      string `json:"avatar"`
	Description string `json:"description"`
	Prize       int    `json:"prize"`
}

func ToGiftDTO(gift *lottery_grpc.Gift) GiftDTO {
	return GiftDTO{
		ID:          gift.ID,
		Name:        gift.Name,
		Avatar:      gift.Avatar,
		Description: gift.Description,
		Prize:       int(gift.Prize),
	}
}

type OrderDTO struct {
	ID        int64            `json:"id,string"` // 订单 ID
	User      userdto.BriefDTO `json:"user"`
	Gift      GiftDTO          `json:"gift"`
	Count     int              `json:"count"`      // 购买数量
	CreatedAt string           `json:"created_at"` // 创建时间
}

func ToOrderDTO(order *lottery_grpc.Order, user *user_grpc.UserDetail) OrderDTO {
	return OrderDTO{
		ID:        order.OrderID,
		User:      userdto.ToBriefDTO(user),
		Gift:      ToGiftDTO(order.Gift),
		Count:     int(order.Count),
		CreatedAt: order.CreatedAt,
	}
}
