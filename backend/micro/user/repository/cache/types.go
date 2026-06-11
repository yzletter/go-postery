package cache

import "context"

type UserCache interface {
	ChangeScore(ctx context.Context, uid int64, delta int) error
	Top(ctx context.Context) ([]int64, []float64, error)
}
