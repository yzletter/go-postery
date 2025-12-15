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
	SetKey(ctx context.Context, pid int64, fields []model.PostCntField, vals []int)
}

type CommentCache interface {
}

type LikeCache interface {
}

type TagCache interface {
}

type FollowCache interface {
}
