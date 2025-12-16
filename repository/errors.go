package repository

import (
	"errors"

	"github.com/yzletter/go-postery/repository/dao"
)

var (
	ErrInternal       = dao.ErrInternal
	ErrRecordNotFound = dao.ErrRecordNotFound
	ErrUniqueKey      = dao.ErrUniqueKey
	ErrParamsInvalid  = dao.ErrParamsInvalid
)

func toRepoErr(err error) error {
	switch {
	case errors.Is(err, dao.ErrInternal):
		return ErrInternal
	case errors.Is(err, dao.ErrRecordNotFound):
		return ErrRecordNotFound
	case errors.Is(err, dao.ErrUniqueKey):
		return ErrUniqueKey
	case errors.Is(err, dao.ErrParamsInvalid):
		return ErrParamsInvalid
	default:
		return ErrInternal
	}
}
