package service

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hashedPassword, plainPassword string) error
}

type IDGenerator interface {
	NextID() int64
}

type JwtManager interface {
	GenToken(claims JwtClaim, expiration int64) (string, error)
	VerifyToken(token string) (*JwtClaim, error)
}

type RefreshSessionStore interface {
}

type JwtClaim struct {
	ID   int64
	Role int
	SSid string
}
