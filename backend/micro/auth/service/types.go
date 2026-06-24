package service

import (
	"context"

	"github.com/yzletter/go-postery/backend/ports"
)

// AuthService 用户认证服务
type AuthService interface {
	// LoginByPassword 手机号码 / 邮箱 + 密码登录
	//
	// Parameter:
	//	- identifier: 登录凭证
	//	- password: 密码
	//
	// Return:
	//	- int64: 用户 ID
	//	- error: 可能返回的错误
	LoginByPassword(ctx context.Context, identifier string, password string) (int64, error)

	// LoginByPhone 手机号码 + 验证码进行登录, 未注册的手机号码自动进行注册
	//
	// Parameter:
	//	- phone: 手机号码
	//	- code: 验证码
	//
	// Return:
	//	- int64: 用户 ID
	//	- error: 可能返回的错误
	LoginByPhone(ctx context.Context, phone string, code string) (int64, error)

	// HasPassword 查询密码状态
	//
	// Parameter:
	//	- id: 用户 ID
	//
	// Return:
	//	- bool: 是否已设置密码
	//	- error: 可能返回的错误
	HasPassword(ctx context.Context, id int64) (bool, error)

	// SetPassword 初始化密码
	//
	// Parameter:
	//	- uid: 用户 ID
	//	- code: 验证码
	//	- password: 密码
	//
	// Return:
	//	- error: 可能返回的错误
	SetPassword(ctx context.Context, uid int64, code string, password string) error

	// UpdatePassword 更新密码
	//
	// Parameter:
	//	- uid: 用户 ID
	//	- oldPassword: 旧密码
	//	- newPassword: 新密码
	//
	// Return:
	//	- error: 可能返回的错误
	UpdatePassword(ctx context.Context, uid int64, oldPassword string, newPassword string) error

	// GetAuthIdentityByUID 获取用户身份认证
	//
	// Parameter:
	//	- id: 用户 ID
	//
	// Return:
	//	- string: 手机号码
	//	- string: 邮箱
	//	- error: 可能返回的错误
	GetAuthIdentityByUID(ctx context.Context, id int64) (string, string, error)

	// IssueTokens 签发双 Token
	//
	// Parameter:
	//	- uid: 用户 ID
	//	- role: 用户角色
	//	- userAgent: 用户代理
	//
	// Return:
	//	- string: AccessToken
	//	- string: RefreshToken
	//	- error: 可能返回的错误
	IssueTokens(ctx context.Context, uid int64, role int, userAgent string) (string, string, error)

	// ClearTokens 清除双 Token
	//
	// Parameter:
	//	- accessToken: AccessToken
	//	- refreshToken: RefreshToken
	//
	// Return:
	//	- error: 可能返回的错误
	ClearTokens(ctx context.Context, accessToken string, refreshToken string) error

	// VerifyAccessToken 校验 AccessToken
	//
	// Parameter:
	//	- accessToken: AccessToken
	//
	// Return:
	//	- *ports.JWTTokenClaims: Token 声明
	//	- error: 可能返回的错误
	VerifyAccessToken(ctx context.Context, accessToken string) (*ports.JWTTokenClaims, error)

	// GetInfoByRefreshToken 根据 RefreshToken 获取用户信息, 用于重新签发双 Token
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

	// CheckBlackList 根据 SSID 检查黑名单, 检查用户是否被拉黑
	//
	// Parameter:
	//	- ssid: 会话 ID
	//
	// Return:
	//	- bool: 是否在黑名单中
	//	- error: 可能返回的错误
	CheckBlackList(ctx context.Context, ssid string) (bool, error)
}
