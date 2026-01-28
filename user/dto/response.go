package dto

import (
	"time"

	user_grpc "github.com/yzletter/go-postery/api/proto/user/v1"
	"github.com/yzletter/go-postery/user/model"
)

// ToTopUser model.UserProfile 转 user_grpc.TopUser
func ToTopUser(userProfile *model.UserProfile, score float64) *user_grpc.TopUser {
	var res = &user_grpc.TopUser{
		ID:       userProfile.UserID,
		Nickname: userProfile.Nickname,
		Bio:      "",
		Avatar:   "",
		Score:    float32(score),
	}

	if userProfile.Bio != nil {
		res.Bio = *userProfile.Bio
	}

	if userProfile.Avatar != nil {
		res.Avatar = *userProfile.Avatar
	}

	return res
}

// ToUserDetail model.UserProfile 转 user_grpc.UserDetail
func ToUserDetail(profile *model.UserProfile) *user_grpc.UserDetail {
	var res = &user_grpc.UserDetail{
		ID:          profile.UserID,
		Nickname:    profile.Nickname,
		Avatar:      "",
		Bio:         "",
		Gender:      uint32(profile.Gender),
		Birthday:    "",
		Location:    "",
		Country:     "",
		LastLoginIP: "",
	}

	if profile.Country != nil {
		res.Country = *profile.Country
	}
	if profile.LastLoginIP != nil {
		res.LastLoginIP = *profile.LastLoginIP
	}
	if profile.Location != nil {
		res.Location = *profile.Location
	}
	if profile.Birthday != nil {
		res.Birthday = profile.Birthday.Format(time.RFC3339)
	}
	if profile.Bio != nil {
		res.Bio = *profile.Bio
	}
	if profile.Avatar != nil {
		res.Avatar = *profile.Avatar
	}
	return res
}

// ToUserBrief model.UserProfile 转 user_grpc.UserBrief
func ToUserBrief(profile *model.UserProfile) *user_grpc.UserBrief {
	var res = &user_grpc.UserBrief{
		ID:       profile.UserID,
		Nickname: profile.Nickname,
		Avatar:   "",
	}

	if profile.Avatar != nil {
		res.Avatar = *profile.Avatar
	}
	return res
}
