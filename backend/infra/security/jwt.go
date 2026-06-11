package security

import (
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/yzletter/go-postery/backend/ports"
)

type jwtManager struct {
	tokenKey []byte
}

func NewJwtManager(tokenKey string) ports.JwtManager {
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

func (manager *jwtManager) GenToken(claim ports.JWTTokenClaims) (string, error) {
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

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwtClaims)
	tokenString, err := token.SignedString(manager.tokenKey)
	if err != nil {
		slog.Error("Token Gen Failed", "error", err)
		return "", ports.ErrTokenGenFailed
	}

	return tokenString, nil
}

func (manager *jwtManager) VerifyToken(tokenString string) (*ports.JWTTokenClaims, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS512 {
			return nil, ports.ErrTokenInvalid
		}
		return manager.tokenKey, nil
	}

	claims := &myJwtClaim{}
	token, err := jwt.ParseWithClaims(tokenString, claims, keyFunc)
	if err != nil || token == nil || !token.Valid {
		return nil, ports.ErrTokenInvalid
	}

	return &ports.JWTTokenClaims{
		Uid:       claims.Uid,
		SSid:      claims.SSid,
		Role:      claims.Role,
		UserAgent: claims.UserAgent,
		Issuer:    claims.Issuer,
		Subject:   claims.Subject,
		Audience:  []string(claims.Audience),
		ExpiresAt: toTimePtr(claims.ExpiresAt),
		NotBefore: toTimePtr(claims.NotBefore),
		IssuedAt:  toTimePtr(claims.IssuedAt),
		ID:        claims.ID,
	}, nil
}

func toNumericDate(t *time.Time) *jwt.NumericDate {
	if t == nil {
		return nil
	}
	return jwt.NewNumericDate(*t)
}

func toTimePtr(nd *jwt.NumericDate) *time.Time {
	if nd == nil {
		return nil
	}
	t := nd.Time
	return &t
}
