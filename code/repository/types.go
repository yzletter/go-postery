package repository

import (
	"context"

	"github.com/yzletter/go-postery/code/model"
)

type CodeRepository interface {
	Allow(ctx context.Context, biz model.CodeBiz, field string, code string) error             // Allow 判断是否允许发送 Code
	CheckCode(ctx context.Context, biz model.CodeBiz, field string, code string) (bool, error) // CheckCode 校验 Code
}
