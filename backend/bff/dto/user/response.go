package user

import user_grpc "github.com/yzletter/go-postery/api/proto/user/v1"

type OSSSignDTO struct {
	Response string `json:"response"`
}

// BriefDTO 后端返回简要 User 信息
type BriefDTO struct {
	ID       int64  `json:"id,string"` // ID
	Nickname string `json:"nickname"`  // 昵称
	Avatar   string `json:"avatar"`    // 头像 URL
	Bio      string `json:"bio"`       // 个性签名
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
	ID       int64  `json:"id,string"`
	Nickname string `json:"nickname"` // 用户名
	Bio      string `json:"bio"`      // 个性签名
	Avatar   string `json:"avatar"`   // 头像 URL
	Score    int64  `json:"score"`
}

// ToBriefDTO user_grpc.Profile 转 BriefDTO
func ToBriefDTO(profile *user_grpc.Profile) BriefDTO {
	var res = BriefDTO{
		ID:       profile.UserID,
		Nickname: profile.Nickname,
		Avatar:   profile.Avatar,
		Bio:      profile.Bio,
	}
	return res
}

// BriefsToDTO []*user_grpc.ProfileBrief 转 []BriefDTO
func BriefsToDTO(briefs []*user_grpc.ProfileBrief) []BriefDTO {
	res := make([]BriefDTO, 0, len(briefs))

	for _, b := range briefs {
		briefDTO := BriefDTO{
			ID:       b.UserID,
			Nickname: b.Nickname,
			Avatar:   b.Avatar,
			Bio:      b.Bio,
		}
		res = append(res, briefDTO)
	}
	return res
}

// ToTopDTO user_grpc.TopResponse 转 ToTopDTO
func ToTopDTO(topUser *user_grpc.ProfileTop) TopDTO {
	return TopDTO{
		ID:       topUser.UserID,
		Nickname: topUser.Nickname,
		Bio:      topUser.Bio,
		Avatar:   topUser.Avatar,
		Score:    topUser.Score,
	}
}

// ToDetailDTO user_grpc.Profile 转 DetailDTO
func ToDetailDTO(profile *user_grpc.Profile) DetailDTO {
	return DetailDTO{
		ID:          profile.UserID,
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
