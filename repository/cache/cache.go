package cache

import (
	"context"

	"github.com/yzletter/go-postery/model"
)

// 定义 Cache 层所有接口

type UserCache interface {
}

type PostCache interface {
	ChangeInteractiveCnt(ctx context.Context, pid int64, field model.PostCntField, delta int) (bool, error)
	SetInteractiveKey(ctx context.Context, pid int64, fields []model.PostCntField, vals []int)
	SetScore(ctx context.Context, pid int64) error
	CheckPostLikeTime(ctx context.Context, pid int64) (float64, error)
	ChangeScore(ctx context.Context, pid int64, delta int) error
	Top(ctx context.Context) ([]int64, []float64, error)
}

type CommentCache interface {
}

type LikeCache interface {
}

type TagCache interface {
}

type FollowCache interface {
}

type MessageCache interface{}
type SessionCache interface{}
