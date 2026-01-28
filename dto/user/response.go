package user

import (
	"time"

	"github.com/yzletter/go-postery/model"
)

// BriefDTO 后端返回简要 User 信息
type BriefDTO struct {
	ID       int64  `json:"id,string"` // ID
	Nickname string `json:"nickname"`  // 昵称
	Avatar   string `json:"avatar"`    // 头像 URL
}

// DetailDTO 后端返回详细 User 信息
type DetailDTO struct {
	ID          int64  `json:"id,string"`     // ID 雪花算法
	Nickname    string `json:"nickname"`      // 用户名
	Avatar      string `json:"avatar"`        // 头像 URL
	Bio         string `json:"bio"`           // 个性签名
	Gender      int    `json:"gender"`        // 性别: 0 表示空, 1 表示男, 2 表示女, 3 表示其它
	BirthDay    string `json:"birthday"`      // 生日
	Location    string `json:"location"`      // 地区
	Country     string `json:"country"`       // 国家
	LastLoginIP string `json:"last_login_ip"` // 最近一次登录 IP
}

// TopDTO 后端返回排行榜 User 信息
type TopDTO struct {
	ID       int64   `json:"id,string"`
	Nickname string  `json:"nickname"` // 用户名
	Bio      string  `json:"bio"`      // 个性签名
	Avatar   string  `json:"avatar"`   // 头像 URL
	Score    float64 `json:"score"`
}

// ToBriefDTO model.UserProfile 转 BriefDTO
func ToBriefDTO(userProfile *model.UserProfile) BriefDTO {
	var res = BriefDTO{
		ID:       userProfile.UserID,
		Nickname: userProfile.Nickname,
		Avatar:   "",
	}

	if userProfile.Avatar != nil {
		res.Avatar = *userProfile.Avatar
	}
	return res
}

// ToTopDTO model.UserProfile 转 ToTopDTO
func ToTopDTO(userProfile *model.UserProfile, score float64) TopDTO {
	var res = TopDTO{
		ID:       userProfile.UserID,
		Nickname: userProfile.Nickname,
		Bio:      "",
		Avatar:   "",
		Score:    score,
	}

	if userProfile.Bio != nil {
		res.Bio = *userProfile.Bio
	}
	if userProfile.Avatar != nil {
		res.Avatar = *userProfile.Avatar
	}

	return res
}

// ToDetailDTO model.UserProfile 转 DetailDTO
func ToDetailDTO(userProfile *model.UserProfile) DetailDTO {
	var res = DetailDTO{
		ID:          userProfile.UserID,
		Nickname:    userProfile.Nickname,
		Gender:      userProfile.Gender,
		Avatar:      "",
		Bio:         "",
		BirthDay:    "",
		Location:    "",
		Country:     "",
		LastLoginIP: "",
	}

	if userProfile.Country != nil {
		res.Country = *userProfile.Country
	}
	if userProfile.LastLoginIP != nil {
		res.LastLoginIP = *userProfile.LastLoginIP
	}
	if userProfile.Location != nil {
		res.Location = *userProfile.Location
	}
	if userProfile.BirthDay != nil {
		res.BirthDay = userProfile.BirthDay.Format(time.RFC3339)
	}
	if userProfile.Bio != nil {
		res.Bio = *userProfile.Bio
	}
	if userProfile.Avatar != nil {
		res.Avatar = *userProfile.Avatar
	}
	return res
}
