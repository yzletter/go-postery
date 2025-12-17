package security

import (
	"log/slog"

	"github.com/golang-jwt/jwt/v5"
	"github.com/yzletter/go-postery/service"
)

type jwtManager struct {
	tokenKey []byte
}

func NewJwtManager(tokenKey string) service.JwtManager {
	return &jwtManager{
		tokenKey: []byte(tokenKey),
	}
}

// GenToken 生成 token
func (manager *jwtManager) GenToken(claim service.JWTTokenClaims) (string, error) {
	// 1. 生成 Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claim)

	// 3. 对 Token 进行加密
	tokenString, err := token.SignedString(manager.tokenKey) // 用长 token 秘钥进行加密
	if err != nil {
		slog.Error("Token Gen Failed", "error", err)
		return "", service.ErrTokenGenFailed
	}

	return tokenString, nil
}

// VerifyToken 校验 JWT
func (manager *jwtManager) VerifyToken(tokenString string) (*service.JWTTokenClaims, error) {
	// 1. 校验用到的 keyFunc 函数
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		return manager.tokenKey, nil
	}

	// 2. 解析 JWT
	claims := &service.JWTTokenClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, keyFunc)
	if err != nil || token == nil || !token.Valid {
		return nil, service.ErrTokenInvalid
	}

	return claims, nil
}
