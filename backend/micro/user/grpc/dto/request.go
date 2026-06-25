package dto

import (
	"errors"
	"strings"
	"time"

	user_grpc "github.com/yzletter/go-postery/api/proto/user/v1"
)

// UpdateProfileRequestToMap user_grpc.UpdateProfileRequest 转更新字段
//
// 只会转换显式传入的 optional 字段，空字符串表示清空可空字段。
func UpdateProfileRequestToMap(request *user_grpc.UpdateProfileRequest) (map[string]any, error) {
	updates := make(map[string]any)

	if request.Nickname != nil {
		nickname := request.GetNickname()
		if strings.TrimSpace(nickname) == "" || len([]rune(nickname)) > 32 {
			return nil, errors.New("invalid nickname")
		}
		updates["nickname"] = nickname
	}
	if request.Avatar != nil {
		if request.GetAvatar() == "" {
			updates["avatar"] = nil
		} else if len(request.GetAvatar()) > 255 {
			return nil, errors.New("invalid avatar")
		} else {
			updates["avatar"] = request.GetAvatar()
		}
	}
	if request.Bio != nil {
		if request.GetBio() == "" {
			updates["bio"] = nil
		} else if len([]rune(request.GetBio())) > 255 {
			return nil, errors.New("invalid bio")
		} else {
			updates["bio"] = request.GetBio()
		}
	}
	if request.Gender != nil {
		gender := int(request.GetGender())
		if gender < 0 || gender > 3 {
			return nil, errors.New("invalid gender")
		}
		updates["gender"] = gender
	}
	if request.Birthday != nil {
		if request.GetBirthday() == "" {
			updates["birthday"] = nil
		} else {
			t, err := time.Parse("2006-01-02", request.GetBirthday())
			if err != nil {
				return nil, err
			}
			updates["birthday"] = t
		}
	}
	if request.Location != nil {
		if request.GetLocation() == "" {
			updates["location"] = nil
		} else if len([]rune(request.GetLocation())) > 64 {
			return nil, errors.New("invalid location")
		} else {
			updates["location"] = request.GetLocation()
		}
	}
	if request.Country != nil {
		if request.GetCountry() == "" {
			updates["country"] = nil
		} else if len([]rune(request.GetCountry())) > 64 {
			return nil, errors.New("invalid country")
		} else {
			updates["country"] = request.GetCountry()
		}
	}

	return updates, nil
}
