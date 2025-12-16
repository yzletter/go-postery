package service

import (
	"errors"

	"github.com/yzletter/go-postery/errno"
	"github.com/yzletter/go-postery/repository"
)

// 将 Repository 层的 error 映射为 Errno 的错误
func toErrnoErr(err error) error {
	switch {
	case errors.Is(err, repository.ErrServerInternal):
		return errno.ErrServerInternal
	case errors.Is(err, repository.ErrNotFound):
		return errno.ErrUserNotFound
	case errors.Is(err, repository.ErrConflict):
		return errno.ErrUserDuplicated
	default:
		return errno.ErrServerInternal
	}
}
