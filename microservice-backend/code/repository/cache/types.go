package cache

import (
	"context"
)

type CodeCache interface {
	Allow(ctx context.Context, biz int, identifier string, code string) (int, error)
	Verify(ctx context.Context, biz int, identifier string, code string) (bool, error)
}
