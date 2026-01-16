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

type LoginByEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required"`
}

type LoginByEmailAndPasswordRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginByPhoneAndPasswordRequest struct {
	Phone    string `json:"phone" binding:"required,len=11"`
	Password string `json:"password" binding:"required"`
}

type RegisterByPhoneRequest struct {
	Phone    string `json:"phone" binding:"required,len=11"`
	Password string `json:"password" binding:"required"`
	Code     string `json:"code" binding:"required"`
	Nickname string `json:"nickname" binding:"required"`
}

type RegisterByEmailRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	Code     string `json:"code" binding:"required"`
	Nickname string `json:"nickname" binding:"required"`
}
