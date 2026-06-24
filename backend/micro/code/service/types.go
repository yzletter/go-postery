package service

import (
	"context"

	"github.com/yzletter/go-postery/backend/micro/code/domain"
)

type CodeService interface {
	// Send 发送验证码
	//
	// Parameter:
	//	- biz: 业务类型
	//	- identifier: 接收验证码的凭证
	//
	// Return:
	//	- error: 可能返回的错误
	Send(ctx context.Context, biz domain.BizType, identifier string) error

	// Verify 校验验证码
	//
	// Parameter:
	//	- biz: 业务类型
	//	- identifier: 接收验证码的凭证
	//	- code: 验证码
	//
	// Return:
	//	- bool: 是否校验通过
	//	- error: 可能返回的错误
	Verify(ctx context.Context, biz domain.BizType, identifier string, code string) (bool, error)
}
