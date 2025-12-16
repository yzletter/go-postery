package repository

import (
	"errors"

	"github.com/yzletter/go-postery/repository/dao"
)

var (
	ErrServerInternal = errors.New("内部错误")
	ErrNotFound       = errors.New("资源不存在")
	ErrConflict       = errors.New("资源冲突")
	ErrInvalidArg     = errors.New("参数非法")
)

func toRepositoryErr(err error) error {
	switch {
	case errors.Is(err, dao.ErrServerInternal):
		return ErrServerInternal
	case errors.Is(err, dao.ErrRecordNotFound):
		return ErrNotFound
	case errors.Is(err, dao.ErrUniqueKey):
		return ErrConflict
	default:
		return ErrServerInternal
	}
}
