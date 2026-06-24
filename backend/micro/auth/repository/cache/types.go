package cache

import (
	"context"
)

// 定义 Cache 层所有接口

type AuthCache interface {
	// DelRefreshToken 缓存中删除 RefreshToken
	//
	// Parameter:
	//	- refreshToken: RefreshToken
	//
	// Return:
	//	- error: 可能返回的错误
	DelRefreshToken(ctx context.Context, refreshToken string) error

	// CheckBlackList 查看 SSID 是否被拉黑
	//
	// Parameter:
	//	- ssid: 会话 ID
	//
	// Return:
	//	- bool: 是否在黑名单中
	//	- error: 可能返回的错误
	CheckBlackList(ctx context.Context, ssid string) (bool, error)

	// GetInfoByRefreshToken 根据 RefreshToken 从缓存中读取用户信息
	//
	// Parameter:
	//	- refreshToken: RefreshToken
	//
	// Return:
	//	- int64: 用户 ID
	//	- int: 用户角色
	//	- string: 用户代理
	//	- error: 可能返回的错误
	GetInfoByRefreshToken(ctx context.Context, refreshToken string) (int64, int, string, error)

	// SetBlackList 拉黑 SSID
	//
	// Parameter:
	//	- ssid: 会话 ID
	//
	// Return:
	//	- error: 可能返回的错误
	SetBlackList(ctx context.Context, ssid string) error

	// SetInfo 根据 RefreshToken 在缓存中存储用户信息
	//
	// Parameter:
	//	- refreshToken: RefreshToken
	//	- mp: 用户信息
	//
	// Return:
	//	- error: 可能返回的错误
	SetInfo(ctx context.Context, refreshToken string, mp map[string]any) error
}
