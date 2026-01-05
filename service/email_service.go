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

type emailService struct {
	emailRepository repository.EmailRepository
	emailManager    ports.EmailManager
}

func NewEmailService(emailRepo repository.EmailRepository, emailManager ports.EmailManager) EmailService {
	return &emailService{
		emailRepository: emailRepo,
		emailManager:    emailManager,
	}
}

func (svc *emailService) SendSMS(ctx context.Context, emailAddress string) error {
	// 生成验证码
	code := svc.generateCode()

	// 写缓存
	err := svc.emailRepository.CheckCode(ctx, emailAddress, code)
	if err != nil {
		// 业务层面错误
		if errors.Is(err, repository.ErrResourceConflict) {
			return errno.ErrSendToFrequent
		}
		// 系统层面错误
		return errno.ErrServerInternal
	}

	// 发送短信
	err = svc.emailManager.Send(emailAddress, code)
	if err != nil {
		// 系统层面错误
		return errno.ErrServerInternal
	}

	return nil
}

func (svc *emailService) CheckSMS(ctx context.Context, emailAddress string, code string) error {
	//TODO implement me
	panic("implement me")
}

// 生成 Code
func (svc *emailService) generateCode() string {
	n := rand.IntN(100000)
	return fmt.Sprintf("%06d", n)
}
