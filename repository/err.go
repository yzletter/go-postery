package repository

import (
	"github.com/yzletter/go-postery/repository/dao"
)

var (
	ErrInternal          = dao.ErrInternal
	ErrRecordNotFound    = dao.ErrRecordNotFound
	ErrUniqueKeyConflict = dao.ErrUniqueKeyConflict
	ErrParamsInvalid     = dao.ErrParamsInvalid
)
