package dto

import (
	"time"

	user_grpc "github.com/yzletter/go-postery/api/proto/user/v1"
	"github.com/yzletter/go-postery/backend/micro/user/domain"
)

// ToProfileTop domain.ProfileTop 转 user_grpc.ProfileTop
func ToProfileTop(profile domain.ProfileTop) *user_grpc.ProfileTop {
	var res = &user_grpc.ProfileTop{
		UserID:   profile.UserID,
		Nickname: profile.Nickname,
		Bio:      profile.Bio,
		Avatar:   profile.Avatar,
		Score:    profile.Score,
	}

	return res
}

// ToProfile domain.Profile 转 user_grpc.Profile
func ToProfile(profile domain.Profile) *user_grpc.Profile {
	var res = &user_grpc.Profile{
		UserID:      profile.UserID,
		Nickname:    profile.Nickname,
		Avatar:      profile.Avatar,
		Bio:         profile.Bio,
		Gender:      int32(profile.Gender),
		Birthday:    "",
		Location:    profile.Location,
		Country:     profile.Country,
		LastLoginAt: "",
		LastLoginIP: profile.LastLoginIP,
		CreatedAt:   profile.CreatedAt.Format(time.RFC3339),
	}

	if !profile.Birthday.IsZero() {
		res.Birthday = profile.Birthday.Format(time.RFC3339)
	}
	if !profile.LastLoginAt.IsZero() {
		res.LastLoginAt = profile.LastLoginAt.Format(time.RFC3339)
	}
	return res
}

// ToProfileBrief domain.ProfileBrief 转 user_grpc.ProfileBrief
func ToProfileBrief(profile domain.ProfileBrief) *user_grpc.ProfileBrief {
	var res = &user_grpc.ProfileBrief{
		UserID:   profile.UserID,
		Nickname: profile.Nickname,
		Avatar:   profile.Avatar,
		Bio:      profile.Bio,
	}

	return res
}
