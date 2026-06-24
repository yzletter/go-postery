package repository

import (
	"context"
	"log/slog"

	"github.com/yzletter/go-postery/backend/micro/code/domain"
	"github.com/yzletter/go-postery/backend/micro/code/model"
	"github.com/yzletter/go-postery/backend/micro/code/repository/cache"
	"github.com/yzletter/go-postery/backend/micro/code/repository/dao"
	"github.com/yzletter/go-postery/backend/micro/code/script"
)

type codeRepository struct {
	dao   dao.CodeDAO
	cache cache.CodeCache
}

func NewCodeRepository(dao dao.CodeDAO, cache cache.CodeCache) CodeRepository {
	return &codeRepository{
		dao:   dao,
		cache: cache,
	}
}

// Allow 判断能否发送验证码
func (repo *codeRepository) Allow(ctx context.Context, biz domain.BizType, identifier string, code string) error {
	res, err := repo.cache.Allow(ctx, biz, identifier, code)
	if err != nil { // Lua 脚本错误
		return ErrServerInternal
	}

	switch res {
	case script.AllowCodeResultAbnormal: // biz 错误或者 redis 中 Key 异常
		return ErrServerInternal
	case script.AllowCodeResultTooFrequent: // 验证码发送过于频繁
		return ErrResourceConflict
	}

	// 发送成功
	return nil
}

// Verify 校验验证码
func (repo *codeRepository) Verify(ctx context.Context, biz domain.BizType, identifier string, code string, codeHash string) (bool, error) {
	res, err := repo.cache.Verify(ctx, biz, identifier, code)
	if err != nil {
		// Lua 脚本错误
		return false, ErrServerInternal
	}

	var ok bool
	switch res {
	case script.VerifyCodeResultNotFound: // 验证码不存在或已过期
		ok = false
	case script.VerifyCodeResultMismatch: // 验证码错误
		ok = false
	case script.VerifyCodeResultMatched: // 验证码正确
		ok = true
	default:
		return false, ErrServerInternal
	}

	if ok {
		if err := repo.dao.MarkVerified(ctx, biz, identifier, codeHash); err != nil {
			slog.Warn("mark code verified failed", "biz", biz, "error", err)
		}
	}
	return ok, nil
}

// RecordSend 验证码落库
func (repo *codeRepository) RecordSend(ctx context.Context, codeRecord domain.CodeRecord) error {
	record := &model.VerificationCode{
		ID:         codeRecord.ID,
		Biz:        int(codeRecord.Biz),
		Identifier: codeRecord.Identifier,
		CodeHash:   codeRecord.CodeHash,
		Status:     model.CodeStatusSent,
		ExpiresAt:  codeRecord.ExpiresAt,
	}
	if err := repo.dao.Create(ctx, record); err != nil {
		return toRepositoryErr(err)
	}
	return nil
}
