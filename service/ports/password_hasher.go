package ports

import "errors"

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hashedPassword, plainPassword string) error
}

// 定义 PasswordHasher 所需要返回的错误
var (
	ErrHashFailed      = errors.New("密码哈希失败")
	ErrInvalidPassword = errors.New("密码错误")
)
