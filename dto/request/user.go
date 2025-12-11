package request

import (
	"time"

	"github.com/yzletter/go-postery/model"
)

// UserInformation 用于存放进 ctx 的用户信息
type UserInformation struct {
	Id   string `json:"id"`
	Name string
}

// LoginUserRequest 定义前端提交登录表单信息的模型映射
type LoginUserRequest struct {
	Name     string `json:"name" form:"name" binding:"required,gte=2"`          // 长度 >= 2
	PassWord string `json:"password" form:"password" binding:"required,len=32"` // 长度 == 32
}

// CreateUserRequest 定义前端提交注册表单信息的模型映射
type CreateUserRequest struct {
	Name     string `json:"name" form:"name" binding:"required,gte=2"`          // 长度 >= 2
	PassWord string `json:"password" form:"password" binding:"required,len=32"` // 长度 == 32
}

// ModifyUserPassRequest 定义前端提交修改密码表单信息的模型映射
type ModifyUserPassRequest struct {
	OldPass string `json:"old_pass" form:"old_pass" binding:"required,len=32"` // 长度 == 32
	NewPass string `json:"new_pass" form:"new_pass" binding:"required,len=32"` // 长度 == 32
}

type ModifyUserProfileRequest struct {
	Email    string `json:"email,omitempty"`    // 邮箱
	Avatar   string `json:"avatar,omitempty"`   // 头像 URL
	Bio      string `json:"bio,omitempty"`      // 个性签名
	Gender   int    `json:"gender,omitempty"`   // 性别: 0 表示空, 1 表示男, 2 表示女, 3 表示其它
	BirthDay string `json:"birthday,omitempty"` // 生日
	Location string `json:"location,omitempty"` // 地区
	Country  string `json:"country,omitempty"`  // 国家
}

func ModifyUserProfileRequestToModel(request ModifyUserProfileRequest) model.User {
	user := model.User{
		Email:    request.Email,
		Avatar:   request.Avatar,
		Bio:      request.Bio,
		Gender:   request.Gender,
		BirthDay: nil,
		Location: request.Location,
		Country:  request.Country,
	}
	if request.BirthDay != "" {
		t, err := time.Parse(time.RFC3339, request.BirthDay)
		if err != nil {
			return user
		}
		user.BirthDay = &t
	}

	return user
}
