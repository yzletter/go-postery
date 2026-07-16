package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"time"

	"github.com/yzletter/go-postery/backend/conf"
	"github.com/yzletter/go-postery/backend/grpc/errs"
	"github.com/yzletter/go-postery/backend/micro/code/domain"
	"github.com/yzletter/go-postery/backend/micro/code/repository"
	"github.com/yzletter/go-postery/backend/ports"
)

// 验证码服务
type codeService struct {
	repository  repository.CodeRepository // Code 模块 Repository 层
	emailClient ports.CodeClient          // 发送邮箱验证码
	smsClient   ports.CodeClient          // 发送短信验证码
	idGen       ports.IDGenerator
}

// NewCodeService 构造函数
func NewCodeService(repository repository.CodeRepository, emailClient ports.CodeClient, smsClient ports.CodeClient, idGen ports.IDGenerator) CodeService {
	return &codeService{
		repository:  repository,
		emailClient: emailClient,
		smsClient:   smsClient,
		idGen:       idGen,
	}
}

// Send 发送验证码并落库
func (svc *codeService) Send(ctx context.Context, biz domain.BizType, identifier string) error {
	// 生成验证码
	code := generateCode()

	// 检查是否能发送验证码
	if err := svc.repository.Allow(ctx, biz, identifier, code); err != nil {
		if errors.Is(err, repository.ErrResourceConflict) {
			slog.Info("send code rejected: too frequent", "biz", biz, "identifier_hash", hashIdentifier(identifier))
			return errs.ErrAlreadyExits
		}
		slog.Error("allow code failed", "biz", biz, "identifier_hash", hashIdentifier(identifier), "error", err)
		return errs.ErrInternal
	}

	// 发送验证码
	switch biz {
	case domain.BizSMS:
		if err := svc.smsClient.Send(ctx, identifier, code); err != nil {
			slog.Error("send sms code failed", "biz", biz, "identifier_hash", hashIdentifier(identifier), "error", err)
			return errs.ErrInternal
		}
	case domain.BizEmail:
		if err := svc.emailClient.Send(ctx, identifier, code); err != nil {
			slog.Error("send email code failed", "biz", biz, "identifier_hash", hashIdentifier(identifier), "error", err)
			return errs.ErrInternal
		}
	default:
		slog.Info("send code rejected: invalid biz", "biz", biz)
		return errs.ErrInvalidArgument
	}

	// 验证码落库
	codeRecord := domain.CodeRecord{
		ID:         svc.idGen.NextID(),
		Biz:        biz,
		Identifier: identifier,
		CodeHash:   hashCode(code),
		ExpiresAt:  expiresAt(biz),
	}
	if err := svc.repository.RecordSend(ctx, codeRecord); err != nil {
		slog.Warn("record code send failed", "biz", biz, "identifier_hash", hashIdentifier(identifier), "error", err)
	}

	return nil
}

// Verify 校验验证码并标识验证码已验证
func (svc *codeService) Verify(ctx context.Context, biz domain.BizType, identifier string, code string) (bool, error) {
	ok, err := svc.repository.Verify(ctx, biz, identifier, code, hashCode(code))
	if err != nil {
		slog.Error("verify code failed", "biz", biz, "identifier_hash", hashIdentifier(identifier), "error", err)
		return false, errs.ErrInternal
	}

	return ok, nil
}

// 生成六位验证码
func generateCode() string {
	n := rand.IntN(1000000)
	return fmt.Sprintf("%06d", n)
}

// 过期时间
func expiresAt(biz domain.BizType) time.Time {
	switch biz {
	case domain.BizSMS:
		return time.Now().Add(time.Duration(conf.SMSValidTime) * time.Second)
	case domain.BizEmail:
		return time.Now().Add(time.Duration(conf.EmailValidTime) * time.Second)
	default:
		return time.Now().Add(time.Duration(conf.EmailValidTime) * time.Second)
	}
}

// 验证码进行哈希处理
func hashCode(code string) string {
	sum := sha256.Sum256([]byte(code))
	return hex.EncodeToString(sum[:])
}

// 业务字段进行哈希处理
func hashIdentifier(identifier string) string {
	sum := sha256.Sum256([]byte(identifier))
	return hex.EncodeToString(sum[:])[:12]
}
