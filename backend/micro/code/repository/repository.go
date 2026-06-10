package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"time"

	"github.com/yzletter/go-postery/backend/conf"
	"github.com/yzletter/go-postery/backend/micro/code/model"
	"github.com/yzletter/go-postery/backend/micro/code/repository/cache"
	"github.com/yzletter/go-postery/backend/micro/code/repository/dao"
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

func (repo *codeRepository) Allow(ctx context.Context, biz int, identifier string, code string) error {
	result, err := repo.cache.Allow(ctx, biz, identifier, code)
	if err != nil { // Lua 脚本错误
		slog.Error("Lua Script Eval Failed", "error", err)
		return ErrServerInternal
	} else if result == -1 { // biz 错误或者 redis 中 Key 异常
		slog.Error("Invalid Biz Or Redis Key")
		return ErrServerInternal
	} else if result == 0 { // 验证码发送过于频繁
		return ErrResourceConflict
	}

	// 发送成功
	return nil
}

func (repo *codeRepository) RecordSend(ctx context.Context, biz int, identifier string, code string) error {
	expireTime, ok := expiresAt(biz)
	if !ok {
		return ErrServerInternal
	}

	record := &model.VerificationCode{
		Biz:        biz,
		Identifier: identifier,
		CodeHash:   hashCode(code),
		Status:     model.CodeStatusSent,
		ExpiresAt:  expireTime,
	}
	if err := repo.dao.Create(ctx, record); err != nil {
		return toRepositoryErr(err)
	}
	return nil
}

func (repo *codeRepository) Verify(ctx context.Context, biz int, identifier string, code string) (bool, error) {
	if ok, err := repo.cache.Verify(ctx, biz, identifier, code); err != nil {
		// Lua 脚本错误
		return false, ErrServerInternal
	} else {
		if ok {
			if err := repo.dao.MarkVerified(ctx, biz, identifier, hashCode(code)); err != nil {
				slog.Error("Mark Code Verified Failed", "biz", biz, "identifier", identifier, "error", err)
			}
		}
		return ok, nil
	}
}

func hashCode(code string) string {
	sum := sha256.Sum256([]byte(code))
	return hex.EncodeToString(sum[:])
}

func expiresAt(biz int) (time.Time, bool) {
	switch biz {
	case 1:
		return time.Now().Add(time.Duration(conf.SMSValidTime) * time.Second), true
	case 2:
		return time.Now().Add(time.Duration(conf.EmailValidTime) * time.Second), true
	default:
		return time.Time{}, false
	}
}
