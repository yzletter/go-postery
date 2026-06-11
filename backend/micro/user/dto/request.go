package dto

import (
	"time"

	user_grpc "github.com/yzletter/go-postery/api/proto/user/v1"
	"github.com/yzletter/go-postery/backend/micro/user/model"
)

// UpdateProfileRequestToModel user_grpc.UpdateProfileRequest 转 model.UserProfile
func UpdateProfileRequestToModel(request *user_grpc.UpdateProfileRequest) *model.UserProfile {
	profile := &model.UserProfile{
		Nickname: request.Nickname,
		Avatar:   &request.Avatar,
		Bio:      &request.Bio,
		Gender:   int(request.Gender),
		Birthday: nil,
		Location: &request.Location,
		Country:  &request.Country,
	}

	t, err := time.Parse("2006-01-02", request.Birthday)
	if err == nil {
		profile.Birthday = &t
	}

	return profile
}
