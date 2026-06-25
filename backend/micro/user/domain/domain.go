package domain

import (
	"time"

	"github.com/yzletter/go-postery/backend/micro/user/model"
)

// Profile 个人资料
type Profile struct {
	UserID      int64     // 用户 ID
	Nickname    string    // 用户昵称
	Gender      int       // 性别 0 空, 1 男, 2 女, 3 其他
	Avatar      string    // 头像 URL
	Bio         string    // 个性签名
	Birthday    time.Time // 生日
	Location    string    // 地区
	Country     string    // 国家
	LastLoginIP string    // 最后登录 IP
	LastLoginAt time.Time // 最后登录时间
	CreatedAt   time.Time // 创建时间
}

// Briefed 将完整用户资料转为简略用户资料
func (p Profile) Briefed() ProfileBrief {
	return ProfileBrief{
		UserID:   p.UserID,
		Nickname: p.Nickname,
		Avatar:   p.Avatar,
		Bio:      p.Bio,
	}
}

// Topped 将用户资料和分数转为排行榜用户资料
func (p Profile) Topped(score int64) ProfileTop {
	return ProfileTop{
		ProfileBrief: p.Briefed(),
		Score:        score,
	}
}

// ProfileBrief 个人资料（简略）
type ProfileBrief struct {
	UserID   int64  // 用户 ID
	Nickname string // 用户昵称
	Avatar   string // 头像 URL
	Bio      string // 个性签名
}

// ProfileTop 个人资料（排行榜）
type ProfileTop struct {
	ProfileBrief
	Score int64 // 排行榜分数
}

// ToModelProfile domain.Profile 转 model.Profile
func ToModelProfile(profile Profile) model.Profile {
	return model.Profile{
		UserID:      profile.UserID,
		Nickname:    profile.Nickname,
		Avatar:      stringPtr(profile.Avatar),
		Bio:         stringPtr(profile.Bio),
		Gender:      profile.Gender,
		Birthday:    timePtr(profile.Birthday),
		Location:    stringPtr(profile.Location),
		Country:     stringPtr(profile.Country),
		LastLoginIP: stringPtr(profile.LastLoginIP),
		LastLoginAt: timePtr(profile.LastLoginAt),
		CreatedAt:   profile.CreatedAt,
	}
}

// ToDomainProfile model.Profile 转 domain.Profile
func ToDomainProfile(profile *model.Profile) Profile {
	res := Profile{
		UserID:    profile.UserID,
		Nickname:  profile.Nickname,
		Gender:    profile.Gender,
		CreatedAt: profile.CreatedAt,
	}

	if profile.Avatar != nil {
		res.Avatar = *profile.Avatar
	}
	if profile.Bio != nil {
		res.Bio = *profile.Bio
	}
	if profile.Birthday != nil {
		res.Birthday = *profile.Birthday
	}
	if profile.Location != nil {
		res.Location = *profile.Location
	}
	if profile.Country != nil {
		res.Country = *profile.Country
	}
	if profile.LastLoginIP != nil {
		res.LastLoginIP = *profile.LastLoginIP
	}
	if profile.LastLoginAt != nil {
		res.LastLoginAt = *profile.LastLoginAt
	}

	return res
}

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func timePtr(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}
