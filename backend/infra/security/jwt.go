package security

import (
	"errors"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrTokenGenFailed = errors.New("jwt token gen failed")
	ErrTokenInvalid   = errors.New("jwt token invalid")
)

type JWTTokenClaims struct {
	Uid       int64
	SSid      string
	Role      int
	UserAgent string

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

type jwtManager struct {
	tokenKey []byte
}

func NewJwtManager(tokenKey string) *jwtManager {
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

func (manager *jwtManager) GenToken(claim JWTTokenClaims) (string, error) {
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
		return "", ErrTokenGenFailed
	}

	return tokenString, nil
}

func (manager *jwtManager) VerifyToken(tokenString string) (*JWTTokenClaims, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS512 {
			return nil, ErrTokenInvalid
		}
		return manager.tokenKey, nil
	}

	claims := &myJwtClaim{}
	token, err := jwt.ParseWithClaims(tokenString, claims, keyFunc)
	if err != nil || token == nil || !token.Valid {
		return nil, ErrTokenInvalid
	}

	return &JWTTokenClaims{
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
