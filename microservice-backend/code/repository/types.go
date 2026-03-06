package repository

import (
	"context"
)

type CodeRepository interface {
	Allow(ctx context.Context, biz int, field string, code string) error          // Allow 判断是否允许发送 Code
	Verify(ctx context.Context, biz int, field string, code string) (bool, error) // Verify 校验 Code
}
