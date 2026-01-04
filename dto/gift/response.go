package gift

import "github.com/yzletter/go-postery/model"

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
