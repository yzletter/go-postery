package service

import (
	"context"

	"github.com/yzletter/go-postery/backend/ports"
)

type AuthService interface {
	LoginByPassword(ctx context.Context, identifier string, password string) (int64, error)
	LoginByPhone(ctx context.Context, phone string, code string) (int64, error)
	HasPassword(ctx context.Context, id int64) (bool, error)
	SetPassword(ctx context.Context, uid int64, code string, password string) error
	UpdatePassword(ctx context.Context, uid int64, oldPassword string, newPassword string) error
	GetAuthIdentityByUID(ctx context.Context, id int64) (string, string, error)
	IssueTokens(ctx context.Context, uid int64, role int, userAgent string) (string, string, error)
	ClearTokens(ctx context.Context, accessToken string, refreshToken string) error
	VerifyAccessToken(ctx context.Context, accessToken string) (*ports.JWTTokenClaims, error)
	GetInfoByRefreshToken(ctx context.Context, refreshToken string) (int64, int, string, error)
	CheckBlackList(ctx context.Context, ssid string) (bool, error)
}
