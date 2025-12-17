package service

import "github.com/golang-jwt/jwt/v5"

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hashedPassword, plainPassword string) error
}

type IDGenerator interface {
	NextID() int64
}

type JwtManager interface {
	GenToken(claim JWTTokenClaims) (string, error)
	VerifyToken(tokenString string) (*JWTTokenClaims, error)
}

// JWTTokenClaims 用于生成 JWT Token 的 Claim
type JWTTokenClaims struct {
	Uid       int64
	SSid      string
	Role      int
	UserAgent string
	jwt.RegisteredClaims
}
