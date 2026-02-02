package ports

import (
	"context"
	"errors"
)

// CodeClient 发送验证码服务需要实现的接口
type CodeClient interface {
	Send(ctx context.Context, identifier string, code string) error
}

var ErrSendCodeFailed = errors.New("发送验证码失败")
