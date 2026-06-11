package dao

import (
	"context"

	"github.com/yzletter/go-postery/backend/micro/code/model"
)

const (
	DeleteFailed = "MySQL DeleteScore Record Failed"
	CreateFailed = "MySQL Create Record Failed"
	FindFailed   = "MySQL Find Record Failed"
	UpdateFailed = "MySQL Update Record Failed"
)

type CodeDAO interface {
	Create(ctx context.Context, code *model.VerificationCode) error
	MarkVerified(ctx context.Context, biz int, identifier string, codeHash string) error
}
