package user

import (
	"time"

	"github.com/yzletter/go-postery/model"
)

// BriefDTO 后端返回简要 User 信息
type BriefDTO struct {
	ID     int64  `json:"id,string"` // ID
	Name   string `json:"name"`      // 用户名
	Avatar string `json:"avatar"`    // 头像 URL
}

// DetailDTO 后端返回详细 User 信息
type DetailDTO struct {
	ID          int64  `json:"id,string"`     // ID 雪花算法
	Name        string `json:"name"`          // 用户名
	Email       string `json:"email"`         // 邮箱
	Avatar      string `json:"avatar"`        // 头像 URL
	Bio         string `json:"bio"`           // 个性签名
	Gender      int    `json:"gender"`        // 性别: 0 表示空, 1 表示男, 2 表示女, 3 表示其它
	BirthDay    string `json:"birthday"`      // 生日
	Location    string `json:"location"`      // 地区
	Country     string `json:"country"`       // 国家
	LastLoginIP string `json:"last_login_ip"` // 最近一次登录 IP
}

// ToBriefDTO model.User 转 UserDTO
func ToBriefDTO(user *model.User) BriefDTO {
	return BriefDTO{
		ID:     user.ID,
		Name:   user.Username,
		Avatar: "", // todo
	}
}

// ToDetailDTO model.User 转 DetailDTO
func ToDetailDTO(user *model.User) DetailDTO {
	userDetailDTO := DetailDTO{
		ID:          user.ID,
		Name:        user.Username,
		Email:       user.Email,
		Avatar:      user.Avatar,
		Bio:         user.Bio,
		Gender:      user.Gender,
		BirthDay:    "",
		Location:    user.Location,
		Country:     user.Country,
		LastLoginIP: user.LastLoginIP,
	}

	if user.BirthDay != nil {
		userDetailDTO.BirthDay = user.BirthDay.Format(time.RFC3339)
	}

	return userDetailDTO
}
