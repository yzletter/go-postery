package ports

import (
	"context"
	"errors"
)

type SmsClient interface {
	SendSms(ctx context.Context, phoneNumber string, code string) error
	CheckSms(ctx context.Context, phoneNumber string, code string) error
}

var (
	ErrInvalidCode    = errors.New("验证码错误")
	ErrSendSMSFailed  = errors.New("核验验证码失败")
	ErrCheckSMSFailed = errors.New("核验验证码失败")
)
