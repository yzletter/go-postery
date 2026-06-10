package dao

import (
	"context"

	"github.com/yzletter/go-postery/backend/micro/code/model"
)

type CodeDAO interface {
	Create(ctx context.Context, code *model.VerificationCode) error
	MarkVerified(ctx context.Context, biz int, identifier string, codeHash string) error
}
