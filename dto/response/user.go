package dto

import (
	"time"

	"github.com/yzletter/go-postery/model"
)

// UserBriefDTO 后端返回简要 User 信息
type UserBriefDTO struct {
	Id     int64  `json:"id,string"` // ID 雪花算法
	Name   string `json:"name"`      // 用户名
	Avatar string `json:"avatar"`    // 头像 URL
}

// UserDetailDTO 后端返回详细 User 信息
type UserDetailDTO struct {
	Id          int64  `json:"id,string"`     // ID 雪花算法
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

// ToUserBriefDTO model.User 转 UserDTO
func ToUserBriefDTO(user model.User) UserBriefDTO {
	return UserBriefDTO{
		Id:     user.ID,
		Name:   user.Username,
		Avatar: "", // todo
	}
}

// ToUserDetailDTO model.User 转 UserDetailDTO
func ToUserDetailDTO(user model.User) UserDetailDTO {
	userDetailDTO := UserDetailDTO{
		Id:          user.ID,
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
