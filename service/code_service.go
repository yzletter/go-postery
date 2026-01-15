package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"

	"github.com/yzletter/go-postery/errno"
	"github.com/yzletter/go-postery/model"
	"github.com/yzletter/go-postery/repository"
	"github.com/yzletter/go-postery/service/ports"
)

type codeService struct {
	codeRepo     repository.CodeRepository
	emailManager ports.EmailManager
	smsClient    ports.SmsClient
}

func NewCodeService(codeRepo repository.CodeRepository, emailManager ports.EmailManager, smsClient ports.SmsClient) CodeService {
	return &codeService{
		codeRepo:     codeRepo,
		emailManager: emailManager,
		smsClient:    smsClient,
	}
}

// SendCode 发送验证码
func (svc *codeService) SendCode(ctx context.Context, biz model.CodeBiz, field string) error {
	// 生成验证码
	code := svc.generateCode()

	// 检查是否能发送
	if err := svc.codeRepo.Allow(ctx, biz, field, code); err != nil {
		if errors.Is(err, repository.ErrResourceConflict) {
			return errno.ErrSendToFrequent
		}
		// todo 日志
		return errno.ErrServerInternal
	}

	// 发送验证码
	switch biz {
	case model.EmailCode:
		if err := svc.emailManager.Send(field, code); err != nil {
			return errno.ErrServerInternal
		}
	case model.SMSCode:
		if err := svc.smsClient.SendSms(ctx, field, code); err != nil {
			return errno.ErrServerInternal
		}
	default:
		return errno.ErrInvalidParam
	}
	return nil
}

// CheckCode 校验验证码
func (svc *codeService) CheckCode(ctx context.Context, biz model.CodeBiz, field string, code string) (bool, error) {
	ok, err := svc.codeRepo.CheckCode(ctx, biz, field, code)
	if err != nil {
		return false, errno.ErrServerInternal
	}
	return ok, nil
}

// 生成 Code
func (svc *codeService) generateCode() string {
	n := rand.IntN(100000)
	return fmt.Sprintf("%06d", n)
}
