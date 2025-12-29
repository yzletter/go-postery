package sms

type SendSMSCodeRequest struct {
	PhoneNumber string `json:"phone_number"`
}

type CheckSMSCodeRequest struct {
	PhoneNumber string `json:"phone_number"`
	Code        string `json:"code"`
}
