package repository

import (
	"errors"

	"github.com/yzletter/go-postery/backend/micro/interactive/repository/dao"
)

var (
	ErrServerInternal   = errors.New("内部错误")
	ErrRecordNotFound   = errors.New("资源不存在")
	ErrUniqueKey        = errors.New("唯一键冲突")
	ErrResourceConflict = errors.New("资源冲突")
	ErrParamsInvalid    = errors.New("参数有误")
)

func toRepositoryErr(err error) error {
	switch {
	case errors.Is(err, dao.ErrServerInternal):
		return ErrServerInternal
	case errors.Is(err, dao.ErrRecordNotFound):
		return ErrRecordNotFound
	case errors.Is(err, dao.ErrUniqueKey):
		return ErrUniqueKey
	case errors.Is(err, dao.ErrParamsInvalid):
		return ErrParamsInvalid
	default:
		return ErrServerInternal
	}
}
