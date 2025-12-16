package security

import (
	"errors"

	"github.com/yzletter/go-postery/errno"
	"github.com/yzletter/go-postery/service"
	"golang.org/x/crypto/bcrypt"
)

type BcryptPasswordHasher struct {
	cost int
}

func NewBcryptPasswordHasher(cost int) service.PasswordHasher {
	if cost == 0 {
		cost = bcrypt.DefaultCost
	}
	return &BcryptPasswordHasher{
		cost: cost,
	}
}

func (hasher *BcryptPasswordHasher) Hash(password string) (string, error) {
	res, err := bcrypt.GenerateFromPassword([]byte(password), hasher.cost)
	if err != nil {
		return "", errno.ErrPasswordEncryptFailed
	}

	return string(res), nil
}

func (hasher *BcryptPasswordHasher) Compare(hashedPassword, plainPassword string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return errno.ErrMismatchedHashAndPassword
		}
		return errno.ErrPasswordEncryptFailed
	}
	return nil
}
