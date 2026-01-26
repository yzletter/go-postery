package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"

	"github.com/yzletter/go-postery/api/proto/code/v1"
	"github.com/yzletter/go-postery/code/model"
	"github.com/yzletter/go-postery/code/repository"
	"github.com/yzletter/go-postery/code/service/ports"
	"github.com/yzletter/go-postery/errno"
)

type CodeService struct {
	repository   repository.CodeRepository
	emailManager ports.EmailManager
	smsClient    ports.SmsClient
	code_grpc.UnimplementedCodeServiceServer
}

func NewCodeService(repository repository.CodeRepository, emailManager ports.EmailManager, smsClient ports.SmsClient) *CodeService {
	return &CodeService{
		repository:                     repository,
		emailManager:                   emailManager,
		smsClient:                      smsClient,
		UnimplementedCodeServiceServer: code_grpc.UnimplementedCodeServiceServer{},
	}
}

// Send 发送验证码
func (svc *CodeService) Send(ctx context.Context, req *code_grpc.SendCodeRequest) (*code_grpc.SendCodeResponse, error) {
	// 生成验证码
	newCode := generateCode()

	// 检查是否能发送
	if err := svc.repository.Allow(ctx, model.CodeBiz(req.Biz), req.Identifier, newCode); err != nil {
		if errors.Is(err, repository.ErrResourceConflict) {
			return nil, errno.ErrSendToFrequent
		}
		// todo 日志
		return nil, errno.ErrServerInternal
	}

	// 发送验证码
	switch req.Biz {
	case int64(model.EmailCode):
		if err := svc.emailManager.Send(req.Identifier, newCode); err != nil {
			return nil, errno.ErrServerInternal
		}
	case int64(model.SMSCode):
		if err := svc.smsClient.SendSms(ctx, req.Identifier, newCode); err != nil {
			return nil, errno.ErrServerInternal
		}
	default:
		return nil, errno.ErrInvalidParam
	}

	return &code_grpc.SendCodeResponse{}, nil
}

// Verify 校验验证码
func (svc *CodeService) Verify(ctx context.Context, req *code_grpc.CheckCodeRequest) (*code_grpc.CheckCodeResponse, error) {
	ok, err := svc.repository.CheckCode(ctx, model.CodeBiz(req.Biz), req.Identifier, req.Code)
	if err != nil {
		return &code_grpc.CheckCodeResponse{Result: false}, errno.ErrServerInternal
	}
	return &code_grpc.CheckCodeResponse{Result: ok}, nil
}

// 生成 Code
func generateCode() string {
	n := rand.IntN(100000)
	return fmt.Sprintf("%06d", n)
}
