package auth

type SendEmailCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type SendSMSCodeRequest struct {
	Phone string `json:"phone" binding:"required,len=11"`
}

type LoginByPhoneRequest struct {
	Phone string `json:"phone" binding:"required,len=11"`
	Code  string `json:"code" binding:"required"`
}

type LoginByPasswordRequest struct {
	Identifier string `json:"identifier" binding:"required"`
	Password   string `json:"password" binding:"required"`
}

// UpdatePassRequest 定义前端提交修改密码表单信息的模型映射
type UpdatePassRequest struct {
	OldPass string `json:"old_password" binding:"required,len=32"` // 长度 == 32
	NewPass string `json:"new_password" binding:"required,len=32"` // 长度 == 32
}

type SetPassRequest struct {
	Phone   string `json:"phone" binding:"required,len=11"`
	Code    string `json:"code" binding:"required"`
	NewPass string `json:"new_password" binding:"required,len=32"` // 长度 == 32
}
