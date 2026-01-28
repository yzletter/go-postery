package user

import (
	"time"

	"github.com/yzletter/go-postery/auth/model"
)

type ModifyProfileRequest struct {
	Nickname string `json:"nickname,omitempty"`
	Avatar   string `json:"avatar,omitempty"`   // 头像 URL
	Bio      string `json:"bio,omitempty"`      // 个性签名
	Gender   int    `json:"gender,omitempty"`   // 性别: 0 表示空, 1 表示男, 2 表示女, 3 表示其它
	BirthDay string `json:"birthday,omitempty"` // 生日
	Location string `json:"location,omitempty"` // 地区
	Country  string `json:"country,omitempty"`  // 国家
}

func ModifyProfileRequestToModel(request ModifyProfileRequest) model.UserProfile {
	userProfile := model.UserProfile{
		Nickname: request.Nickname,
		Avatar:   &request.Avatar,
		Bio:      &request.Bio,
		Gender:   request.Gender,
		BirthDay: nil,
		Location: &request.Location,
		Country:  &request.Country,
	}

	t, err := time.Parse("2006-01-02", request.BirthDay)
	if err == nil {
		userProfile.BirthDay = &t
	}

	return userProfile
}
