package security

import "errors"

var (
	ErrMismatchedHashAndPassword = errors.New("密码校验错误")
	ErrHashFailed                = errors.New("密码校验失败")
)
