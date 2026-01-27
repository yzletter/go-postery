package cache

import (
	"context"

	"github.com/yzletter/go-postery/code/model"
)

type CodeCache interface {
	Allow(ctx context.Context, biz model.CodeBiz, field string, code string) (int, error)
	CheckCode(ctx context.Context, biz model.CodeBiz, field string, code string) (bool, error)
}
