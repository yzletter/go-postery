package user

import (
	"github.com/yzletter/go-postery/model"
)

// ModifyPassRequest 定义前端提交修改密码表单信息的模型映射
type ModifyPassRequest struct {
	OldPass string `json:"old_password"  binding:"required,len=32"` // 长度 == 32
	NewPass string `json:"new_password" binding:"required,len=32"`  // 长度 == 32
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
	user := model.User{}
	return user
}
