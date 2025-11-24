package model

// LoginRequest 定义前端提交登录表单信息的模型映射
type LoginRequest struct {
	Name     string `form:"name" binding:"required,gte=2"`      // 长度 >= 2
	PassWord string `form:"password" binding:"required,len=32"` // 长度 == 32
}

// ModifyPasswordRequest 定义前端提交修改密码表单信息的模型映射
type ModifyPasswordRequest struct {
	OldPass string `form:"old_pass" binding:"required,len=32"` // 长度 == 32
	NewPass string `form:"new_pass" binding:"required,len=32"` // 长度 == 32
}
