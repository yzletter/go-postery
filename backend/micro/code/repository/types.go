package repository

import (
	"context"

	"github.com/yzletter/go-postery/backend/micro/code/domain"
)

type CodeRepository interface {
	// Allow 判断是否允许发送 Code
	//
	// Parameter:
	//	- biz: 业务类型
	//	- field: 接收验证码的凭证
	//	- code: 验证码
	//
	// Return:
	//	- error: 可能返回的错误
	Allow(ctx context.Context, biz domain.BizType, field string, code string) error

	// RecordSend 记录已发送 Code
	//
	// Parameter:
	//	- biz: 业务类型
	//	- field: 接收验证码的凭证
	//	- code: 验证码
	//
	// Return:
	//	- error: 可能返回的错误
	RecordSend(ctx context.Context, code domain.CodeRecord) error

	// Verify 校验 Code
	//
	// Parameter:
	//	- biz: 业务类型
	//	- field: 接收验证码的凭证
	//	- code: 验证码
	//	- codeHash: 验证码哈希结果
	//
	// Return:
	//	- bool: 是否校验通过
	//	- error: 可能返回的错误
	Verify(ctx context.Context, biz domain.BizType, identifier string, code string, codeHash string) (bool, error)
}
