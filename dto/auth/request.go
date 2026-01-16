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
