package dao

import (
	"context"

	"github.com/yzletter/go-postery/backend/micro/code/domain"
	"github.com/yzletter/go-postery/backend/micro/code/model"
)

type CodeDAO interface {
	// Create 创建验证码记录
	//
	// Parameter:
	//	- code: 验证码记录
	//
	// Return:
	//	- error: 可能返回的错误
	Create(ctx context.Context, code *model.VerificationCode) error

	// MarkVerified 标记验证码已验证
	//
	// Parameter:
	//	- biz: 业务类型
	//	- identifier: 接收验证码的凭证
	//	- codeHash: 验证码哈希
	//
	// Return:
	//	- error: 可能返回的错误
	MarkVerified(ctx context.Context, biz domain.BizType, identifier string, codeHash string) error
}
