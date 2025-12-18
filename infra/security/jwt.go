package security

import (
	"log/slog"
	"time"

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

type myJwtClaim struct {
	Uid       int64
	SSid      string
	Role      int
	UserAgent string
	jwt.RegisteredClaims
}

// GenToken 生成 token
func (manager *jwtManager) GenToken(claim service.JWTTokenClaims) (string, error) {

	jwtClaims := myJwtClaim{
		Uid:       claim.Uid,
		SSid:      claim.SSid,
		Role:      claim.Role,
		UserAgent: claim.UserAgent,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    claim.Issuer,
			Subject:   claim.Subject,
			Audience:  jwt.ClaimStrings(claim.Audience),
			ExpiresAt: toNumericDate(claim.ExpiresAt),
			NotBefore: toNumericDate(claim.NotBefore),
			IssuedAt:  toNumericDate(claim.IssuedAt),
			ID:        claim.ID,
		},
	}
	// 1. 生成 Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwtClaims)

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
		if token.Method != jwt.SigningMethodHS512 {
			return nil, service.ErrTokenInvalid
		}
		return manager.tokenKey, nil
	}

	// 2. 解析 JWT
	claims := &myJwtClaim{}
	token, err := jwt.ParseWithClaims(tokenString, claims, keyFunc)
	if err != nil || token == nil || !token.Valid {
		return nil, service.ErrTokenInvalid
	}

	aud := []string(claims.Audience)
	res := &service.JWTTokenClaims{
		Uid:       claims.Uid,
		SSid:      claims.SSid,
		Role:      claims.Role,
		UserAgent: claims.UserAgent,
		Issuer:    claims.Issuer,
		Subject:   claims.Subject,
		Audience:  aud,
		ExpiresAt: toTimePtr(claims.ExpiresAt),
		NotBefore: toTimePtr(claims.NotBefore),
		IssuedAt:  toTimePtr(claims.IssuedAt),
		ID:        claims.ID,
	}

	return res, nil
}

// 把 *time.Time 转成 *jwt.NumericDate, 注意判空
func toNumericDate(t *time.Time) *jwt.NumericDate {
	if t == nil {
		return nil
	}
	return jwt.NewNumericDate(*t)
}

// 把 *jwt.NumericDate 转成 *time.Time, 注意判空
func toTimePtr(nd *jwt.NumericDate) *time.Time {
	if nd == nil {
		return nil
	}
	t := nd.Time
	return &t
}
