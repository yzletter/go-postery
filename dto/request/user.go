package request

// UserInformation 用于存放进 ctx 的用户信息
type UserInformation struct {
	Id   int
	Name string
}

// LoginRequest 定义前端提交登录表单信息的模型映射
type LoginRequest struct {
	Name     string `json:"name" form:"name" binding:"required,gte=2"`          // 长度 >= 2
	PassWord string `json:"password" form:"password" binding:"required,len=32"` // 长度 == 32
}

// RegisterRequest 定义前端提交注册表单信息的模型映射
type RegisterRequest struct {
	Name     string `json:"name" form:"name" binding:"required,gte=2"`          // 长度 >= 2
	PassWord string `json:"password" form:"password" binding:"required,len=32"` // 长度 == 32
}

// ModifyPasswordRequest 定义前端提交修改密码表单信息的模型映射
type ModifyPasswordRequest struct {
	OldPass string `json:"old_pass" form:"old_pass" binding:"required,len=32"` // 长度 == 32
	NewPass string `json:"new_pass" form:"new_pass" binding:"required,len=32"` // 长度 == 32
}
