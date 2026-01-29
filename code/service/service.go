package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"

	"github.com/yzletter/go-postery/api/proto/code/v1"
	"github.com/yzletter/go-postery/code/errs"
	"github.com/yzletter/go-postery/code/model"
	"github.com/yzletter/go-postery/code/repository"
	"github.com/yzletter/go-postery/code/service/ports"
)

type codeService struct {
	repository   repository.CodeRepository
	emailManager ports.EmailManager
	smsClient    ports.SmsClient
	code_grpc.UnimplementedCodeServiceServer
}

func NewCodeService(repository repository.CodeRepository, emailManager ports.EmailManager, smsClient ports.SmsClient) CodeService {
	return &codeService{
		repository:                     repository,
		emailManager:                   emailManager,
		smsClient:                      smsClient,
		UnimplementedCodeServiceServer: code_grpc.UnimplementedCodeServiceServer{},
	}
}

// Send 发送验证码
func (svc *codeService) Send(ctx context.Context, req *code_grpc.SendCodeRequest) (*code_grpc.SendCodeResponse, error) {
	// 生成验证码
	newCode := generateCode()

	// 检查是否能发送
	if err := svc.repository.Allow(ctx, model.CodeBiz(req.Biz), req.Identifier, newCode); err != nil {
		if errors.Is(err, repository.ErrResourceConflict) {
			slog.Error("Send Code Too Frequent", "error", err)
			return nil, errs.ErrAlreadyExits
		}
		slog.Error("Server Internal Error", "error", err)
		return nil, errs.ErrInternal
	}

	// 发送验证码
	switch req.Biz {
	case int64(model.EmailCode):
		if err := svc.emailManager.Send(req.Identifier, newCode); err != nil {
			slog.Error("Server Internal Error", "error", err)
			return nil, errs.ErrInternal
		}
	case int64(model.SMSCode):
		if err := svc.smsClient.SendSms(ctx, req.Identifier, newCode); err != nil {
			slog.Error("Server Internal Error", "error", err)
			return nil, errs.ErrInternal
		}
	default:
		slog.Error("Invalid Biz Code")
		return nil, errs.ErrInvalidArgument
	}

	return &code_grpc.SendCodeResponse{}, nil
}

// Verify 校验验证码
func (svc *codeService) Verify(ctx context.Context, req *code_grpc.CheckCodeRequest) (*code_grpc.CheckCodeResponse, error) {
	ok, err := svc.repository.CheckCode(ctx, model.CodeBiz(req.Biz), req.Identifier, req.Code)
	if err != nil {
		slog.Error("Server Internal Error", "error", err)
		return &code_grpc.CheckCodeResponse{Result: false}, errs.ErrInternal
	}
	return &code_grpc.CheckCodeResponse{Result: ok}, nil
}

// 生成 Code
func generateCode() string {
	n := rand.IntN(100000)
	return fmt.Sprintf("%06d", n)
}
