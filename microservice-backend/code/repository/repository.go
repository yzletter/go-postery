package repository

import (
	"context"
	"log/slog"

	"github.com/yzletter/go-postery/microservice-backend/code/repository/cache"
)

type codeRepository struct {
	cache cache.CodeCache
}

func NewCodeRepository(cache cache.CodeCache) CodeRepository {
	return &codeRepository{
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

func (repo *codeRepository) Verify(ctx context.Context, biz int, identifier string, code string) (bool, error) {
	if ok, err := repo.cache.Verify(ctx, biz, identifier, code); err != nil {
		// Lua 脚本错误
		return false, ErrServerInternal
	} else {
		return ok, nil
	}
}
