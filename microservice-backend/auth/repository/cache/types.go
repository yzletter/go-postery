package cache

import (
	"context"
)

// 定义 Cache 层所有接口

type AuthCache interface {
	DelRefreshToken(ctx context.Context, refreshToken string) error
	CheckBlackList(ctx context.Context, ssid string) (bool, error)
	GetInfoByRefreshToken(ctx context.Context, refreshToken string) (int64, int, string, error)
	SetBlackList(ctx context.Context, ssid string) error
	SetInfo(ctx context.Context, refreshToken string, mp map[string]any) error
}
