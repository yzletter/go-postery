package user

import (
	user_grpc "github.com/yzletter/go-postery/api/proto/user/v1"
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

// ToBriefDTO user_grpc.UserDetail 转 BriefDTO
func ToBriefDTO(profile *user_grpc.UserDetail) BriefDTO {
	var res = BriefDTO{
		ID:       profile.ID,
		Nickname: profile.Nickname,
		Avatar:   profile.Avatar,
	}
	return res
}

// BriefsToDTO []*user_grpc.UserBrief 转 []BriefDTO
func BriefsToDTO(briefs []*user_grpc.UserBrief) []BriefDTO {
	res := make([]BriefDTO, 0, len(briefs))

	for _, b := range briefs {
		briefDTO := BriefDTO{
			ID:       b.ID,
			Nickname: b.Nickname,
			Avatar:   b.Avatar,
		}
		res = append(res, briefDTO)
	}
	return res
}

// ToTopDTO user_grpc.TopResponse 转 ToTopDTO
func ToTopDTO(topUser *user_grpc.TopUser) TopDTO {
	return TopDTO{
		ID:       topUser.ID,
		Nickname: topUser.Nickname,
		Bio:      topUser.Bio,
		Avatar:   topUser.Avatar,
		Score:    float64(topUser.Score),
	}
}

// ToDetailDTO user_grpc.UserDetail 转 DetailDTO
func ToDetailDTO(profile *user_grpc.UserDetail) DetailDTO {
	return DetailDTO{
		ID:          profile.ID,
		Nickname:    profile.Nickname,
		Gender:      int(profile.Gender),
		Avatar:      profile.Avatar,
		Bio:         profile.Bio,
		BirthDay:    profile.Birthday,
		Location:    profile.Location,
		Country:     profile.Country,
		LastLoginIP: profile.LastLoginIP,
	}
}
