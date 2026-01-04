package order

import (
	"time"

	giftdto "github.com/yzletter/go-postery/dto/gift"
	userdto "github.com/yzletter/go-postery/dto/user"
	"github.com/yzletter/go-postery/model"
)

type DTO struct {
	ID        int64            `json:"id,string"` // 订单 ID
	User      userdto.BriefDTO `json:"user"`
	Gift      giftdto.DTO      `json:"gift"`
	Count     int              `json:"count"`      // 购买数量
	CreatedAt string           `json:"created_at"` // 创建时间
}

func ToDTO(order *model.Order, user *model.User, gift *model.Gift) DTO {
	return DTO{
		ID:        order.ID,
		User:      userdto.ToBriefDTO(user),
		Gift:      giftdto.ToDTO(gift),
		Count:     order.Count,
		CreatedAt: order.CreatedAt.Format(time.RFC3339),
	}
}
