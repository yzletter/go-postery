package lottery

import (
	"time"

	model2 "github.com/yzletter/go-postery/auth/model"
	userdto "github.com/yzletter/go-postery/bff/dto/user"
	"github.com/yzletter/go-postery/lottery/model"
)

type DTO struct {
	ID          int64  `json:"id,string"`
	Name        string `json:"name"`
	Avatar      string `json:"avatar"`
	Description string `json:"description"`
	Prize       int    `json:"prize"`
}

func ToDTO(gift *model.Gift) DTO {
	return DTO{
		ID:          gift.ID,
		Name:        gift.Name,
		Avatar:      gift.Avatar,
		Description: gift.Description,
		Prize:       gift.Prize,
	}
}

type DTO struct {
	ID        int64            `json:"id,string"` // 订单 ID
	User      userdto.BriefDTO `json:"user"`
	Gift      giftdto.DTO      `json:"gift"`
	Count     int              `json:"count"`      // 购买数量
	CreatedAt string           `json:"created_at"` // 创建时间
}

func ToDTO(order *model.Order, userProfile *model2.UserProfile, gift *model.Gift) DTO {
	return DTO{
		ID:        order.ID,
		User:      userdto.ToBriefDTO(userProfile),
		Gift:      gift.ToDTO(gift),
		Count:     order.Count,
		CreatedAt: order.CreatedAt.Format(time.RFC3339),
	}
}
