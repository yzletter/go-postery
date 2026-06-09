package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"

	"github.com/yzletter/go-postery/microservice-backend/code/errs"
	"github.com/yzletter/go-postery/microservice-backend/code/repository"
	"github.com/yzletter/go-postery/microservice-backend/code/service/ports"
)

// 验证码服务
type codeService struct {
	repository  repository.CodeRepository // Code 模块 Repository 层
	emailClient ports.CodeClient          // 发送邮箱验证码
	smsClient   ports.CodeClient          // 发送短信验证码
}

// NewCodeService 构造函数
func NewCodeService(repository repository.CodeRepository, emailClient ports.CodeClient, smsClient ports.CodeClient) CodeService {
	return &codeService{
		repository:  repository,
		emailClient: emailClient,
		smsClient:   smsClient,
	}
}

// Send 发送验证码
//
// biz 表示业务种类 : 1 表示短信验证码, 2 表示邮箱验证码
//
// identifier 表示业务字段 : 如手机号或邮箱地址
func (svc *codeService) Send(ctx context.Context, biz int, identifier string) error {
	// 生成验证码
	code := generateCode()

	// 检查是否能发送验证码
	if err := svc.repository.Allow(ctx, biz, identifier, code); err != nil {
		if errors.Is(err, repository.ErrResourceConflict) {
			slog.Error("Send Code Too Frequent", "error", err)
			return errs.ErrAlreadyExits
		}
		slog.Error("Server Internal Error", "error", err)
		return errs.ErrInternal
	}

	// 发送验证码
	switch biz {
	case SmsCode:
		if err := svc.smsClient.Send(ctx, identifier, code); err != nil {
			slog.Error("Server Internal Error", "error", err)
			return errs.ErrInternal
		}
	case EmailCode:
		if err := svc.emailClient.Send(ctx, identifier, code); err != nil {
			slog.Error("Server Internal Error", "error", err)
			return errs.ErrInternal
		}
	default:
		slog.Error("Invalid Biz Code")
		return errs.ErrInvalidArgument
	}

	slog.Info("Send Code Success", "Biz", biz, "Identifier", identifier, "Code", code)
	return nil
}

// Verify 校验验证码
//
// biz 表示业务种类 : 1 表示短信验证码, 2 表示邮箱验证码
//
// identifier 表示业务字段 : 如手机号或邮箱地址
//
// code 表示需要校验的验证码
func (svc *codeService) Verify(ctx context.Context, biz int, identifier string, code string) (bool, error) {
	if ok, err := svc.repository.Verify(ctx, biz, identifier, code); err != nil {
		slog.Error("Server Internal Error", "error", err)
		return false, errs.ErrInternal
	} else {
		slog.Info("Verify Code Success", "Biz", biz, "Identifier", identifier, "Code", code, "Result", ok)
		return ok, nil
	}
}

// 生成六位验证码
func generateCode() string {
	n := rand.IntN(100000)
	return fmt.Sprintf("%06d", n)
}
