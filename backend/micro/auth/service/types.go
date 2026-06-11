package service

import (
	"context"

	"github.com/yzletter/go-postery/backend/ports"
)

// AuthService 用户认证服务
type AuthService interface {
	// LoginByPassword 手机号码 / 邮箱 + 密码登录
	LoginByPassword(ctx context.Context, identifier string, password string) (int64, error)

	// LoginByPhone 手机号码 + 验证码进行登录, 未注册的手机号码自动进行注册
	LoginByPhone(ctx context.Context, phone string, code string) (int64, error)

	// HasPassword 查询密码状态
	HasPassword(ctx context.Context, id int64) (bool, error)

	// SetPassword 初始化密码
	SetPassword(ctx context.Context, uid int64, code string, password string) error

	// UpdatePassword 更新密码
	UpdatePassword(ctx context.Context, uid int64, oldPassword string, newPassword string) error

	// GetAuthIdentityByUID 获取用户身份认证
	GetAuthIdentityByUID(ctx context.Context, id int64) (string, string, error)

	// IssueTokens 签发双 Token
	IssueTokens(ctx context.Context, uid int64, role int, userAgent string) (string, string, error)

	// ClearTokens 清除双 Token
	ClearTokens(ctx context.Context, accessToken string, refreshToken string) error

	// VerifyAccessToken 校验 AccessToken
	VerifyAccessToken(ctx context.Context, accessToken string) (*ports.JWTTokenClaims, error)

	// GetInfoByRefreshToken 根据 RefreshToken 获取用户信息, 用于重新签发双 Token
	GetInfoByRefreshToken(ctx context.Context, refreshToken string) (int64, int, string, error)

	// CheckBlackList 根据 SSID 检查黑名单, 检查用户是否被拉黑
	CheckBlackList(ctx context.Context, ssid string) (bool, error)
}
