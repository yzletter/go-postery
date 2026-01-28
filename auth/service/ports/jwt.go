package ports

import (
	"errors"
	"time"
)

// JWTTokenClaims 定义用于生成 JWT Token 的 Claim
type JWTTokenClaims struct {
	// 用户字段
	Uid       int64
	SSid      string
	Role      int
	UserAgent string

	// JWT 字段
	Issuer    string
	Subject   string
	Audience  []string
	ExpiresAt *time.Time
	NotBefore *time.Time
	IssuedAt  *time.Time
	ID        string
}

type JwtManager interface {
	GenToken(claim JWTTokenClaims) (string, error)
	VerifyToken(tokenString string) (*JWTTokenClaims, error)
}

// 定义 JwtManager 所需要返回的错误
var (
	ErrTokenGenFailed = errors.New("JWT Token 生成失败")
	ErrTokenInvalid   = errors.New("JWT Token 不合法")
)
