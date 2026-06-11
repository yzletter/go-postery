package cache

import (
	"context"

	"github.com/yzletter/go-postery/backend/micro/post/model"
)

type PostCache interface {
	ChangeInteractiveCnt(ctx context.Context, pid int64, field model.PostCntField, delta int) (bool, error)
	SetInteractiveKey(ctx context.Context, pid int64, fields []model.PostCntField, vals []int)
	SetScore(ctx context.Context, pid int64) error
	CheckPostLikeTime(ctx context.Context, pid int64) (float64, error)
	ChangeScore(ctx context.Context, pid int64, delta int) error
	Top(ctx context.Context) ([]int64, []float64, error)
	DeleteScore(ctx context.Context, id int64) error
}
