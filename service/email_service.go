package service

import (
	"context"

	"github.com/yzletter/go-postery/repository"
	"github.com/yzletter/go-postery/service/ports"
)

type emailService struct {
	emailRepo    repository.EmailRepository
	emailManager ports.EmailManager
}

func NewEmailService(emailRepo repository.EmailRepository, emailManager ports.EmailManager) EmailService {
	return &emailService{
		emailRepo:    emailRepo,
		emailManager: emailManager,
	}
}

func (svc *emailService) SendSMS(ctx context.Context, emailAddress string) error {
	//TODO implement me
	panic("implement me")
}

func (svc *emailService) CheckSMS(ctx context.Context, emailAddress string, code string) error {
	//TODO implement me
	panic("implement me")
}
