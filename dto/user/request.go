package user

import (
	"time"

	"github.com/yzletter/go-postery/model"
)

// JWTInfo 用于存放进 JWT 自定义字段和放入 ctx 的用户信息
type JWTInfo struct {
	Id   string `json:"user_id"`
	Name string `json:"user_name"`
	Role string `json:"role"`
}

// RegisterRequest 定义前端提交注册表单信息的模型映射
type RegisterRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name" binding:"required,gte=2"`       // 长度 >= 2
	PassWord string `json:"password"  binding:"required,len=32"` // 长度 == 32
}

// LoginRequest 定义前端提交登录表单信息的模型映射
type LoginRequest struct {
	//Email    string `json:"email" binding:"required,gte=2"`
	Name     string `json:"name"  binding:"required,gte=2"`      // 长度 >= 2
	PassWord string `json:"password"  binding:"required,len=32"` // 长度 == 32
}

// ModifyPassRequest 定义前端提交修改密码表单信息的模型映射
type ModifyPassRequest struct {
	OldPass string `json:"old_pass"  binding:"required,len=32"` // 长度 == 32
	NewPass string `json:"new_pass" binding:"required,len=32"`  // 长度 == 32
}

type ModifyProfileRequest struct {
	Email    string `json:"email,omitempty"`    // 邮箱
	Avatar   string `json:"avatar,omitempty"`   // 头像 URL
	Bio      string `json:"bio,omitempty"`      // 个性签名
	Gender   int    `json:"gender,omitempty"`   // 性别: 0 表示空, 1 表示男, 2 表示女, 3 表示其它
	BirthDay string `json:"birthday,omitempty"` // 生日
	Location string `json:"location,omitempty"` // 地区
	Country  string `json:"country,omitempty"`  // 国家
}

func ModifyProfileRequestToModel(request ModifyProfileRequest) model.User {
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
		t, err := time.Parse("2006-01-02", request.BirthDay)
		if err != nil {
			return user
		}
		user.BirthDay = &t
	}

	return user
}
