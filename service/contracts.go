package service

import (
	"errors"
	"time"
)

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hashedPassword, plainPassword string) error
}

// 定义 PasswordHasher 所需要返回的错误
var (
	ErrHashFailed      = errors.New("密码哈希失败")
	ErrInvalidPassword = errors.New("密码错误")
)

type IDGenerator interface {
	NextID() int64
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
