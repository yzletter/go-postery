package service

import (
	"context"
)

type CodeService interface {
	Send(ctx context.Context, biz int, identifier string) error                        // 发送验证码
	Verify(ctx context.Context, biz int, identifier string, code string) (bool, error) // 校验验证码
}
