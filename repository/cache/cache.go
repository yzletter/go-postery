package cache

import (
	"context"

	"github.com/yzletter/go-postery/model"
)

// 定义 Cache 层所有接口
type OrderCache interface {
	CreateTempOrder(ctx context.Context, uid, gid int64) error
	DeleteTempOrder(ctx context.Context, uid int64) error
	GetTempOrderID(ctx context.Context, uid int64) (int64, error)
}

type GiftCache interface {
	InitInventory(ctx context.Context, gifts []*model.Gift)
	GetAllInventory(ctx context.Context) ([]*model.Gift, error)
	ReduceInventory(ctx context.Context, gid int64) error
	IncreaseInventory(ctx context.Context, gid int64) error
}

type AuthCache interface {
	DelRefreshToken(ctx context.Context, refreshToken string) error
	CheckBlackList(ctx context.Context, ssid string) (bool, error)
	GetInfoByRefreshToken(ctx context.Context, refreshToken string) (int64, int, string, error)
	SetBlackList(ctx context.Context, ssid string) error
	SetInfo(ctx context.Context, refreshToken string, mp map[string]any) error
}

type AgentCache interface {
}
