package security

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrHashFailed      = errors.New("hash failed")
	ErrInvalidPassword = errors.New("invalid password")
)

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hashedPassword, plainPassword string) error
}

type BcryptPasswordHasher struct {
	cost int
}

func NewBcryptPasswordHasher(cost int) *BcryptPasswordHasher {
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
		return "", ErrHashFailed
	}

	return string(res), nil
}

func (hasher *BcryptPasswordHasher) Compare(hashedPassword, plainPassword string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrInvalidPassword
		}
		return ErrHashFailed
	}
	return nil
}
