package auth

type LoginByPhone struct {
	Phone string `json:"phone" binding:"required,len=11"`
	Code  string `json:"code" binding:"required"`
}

type LoginByEmail struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required"`
}

type SendEmailCodeRequest struct {
	Email string `json:"email"`
}

type SendSMSCodeRequest struct {
	Phone string `json:"phone"`
}
