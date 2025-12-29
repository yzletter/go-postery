package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"

	"github.com/yzletter/go-postery/errno"
	"github.com/yzletter/go-postery/repository"
	"github.com/yzletter/go-postery/service/ports"
)

type smsService struct {
	smsClient     ports.SmsClient
	smsRepository repository.SmsRepository
}

// 构造函数
func NewSmsService(smsClient ports.SmsClient, smsRepository repository.SmsRepository) SmsService {
	return &smsService{
		smsClient:     smsClient,
		smsRepository: smsRepository,
	}
}

// SendSMS 发送短信验证码
func (svc *smsService) SendSMS(ctx context.Context, phoneNumber string) error {
	// 生成验证码
	code := svc.generateCode()

	// 写缓存
	err := svc.smsRepository.CheckCode(ctx, phoneNumber, code)
	if err != nil {
		// 业务层面错误
		if errors.Is(err, repository.ErrResourceConflict) {
			return errno.ErrSendToFrequent
		}
		// 系统层面错误
		return errno.ErrServerInternal
	}

	// 发送短信
	err = svc.smsClient.SendSms(ctx, phoneNumber, code)
	if err != nil {
		// 系统层面错误
		return errno.ErrServerInternal
	}

	return nil
}

// CheckSMS 检查短信验证码
func (svc *smsService) CheckSMS(ctx context.Context, phoneNumber string, code string) error {
	err := svc.smsClient.CheckSms(ctx, phoneNumber, code)
	if err != nil {
		// 业务层面错误
		if errors.Is(err, ports.ErrInvalidCode) {
			return errno.ErrInvalidSMSCode
		}
		// 系统层面错误
		return errno.ErrServerInternal
	}
	return nil
}

// 生成 Code
func (svc *smsService) generateCode() string {
	n := rand.IntN(100000)
	return fmt.Sprintf("%06d", n)
}
