package dao

import (
	"context"

	"github.com/yzletter/go-postery/micro-backend/code/model"
)

type CodeDAO interface {
	Create(ctx context.Context, code *model.VerificationCode) error
	MarkVerified(ctx context.Context, biz int, identifier string, codeHash string) error
}
