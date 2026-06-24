package repository

import (
	"errors"

	"github.com/yzletter/go-postery/backend/micro/auth/repository/cache"
	"github.com/yzletter/go-postery/backend/micro/auth/repository/dao"
)

var (
	ErrServerInternal   = errors.New("内部错误")
	ErrRecordNotFound   = errors.New("资源不存在")
	ErrInvalidToken     = errors.New("Token 不合法")
	ErrUniqueKey        = errors.New("唯一键冲突")
	ErrResourceConflict = errors.New("资源冲突")
)

// toRepositoryErr 将 DAO / Cache 层错误转换为 Repository 层错误
func toRepositoryErr(err error) error {
	switch {
	case errors.Is(err, dao.ErrServerInternal):
		return ErrServerInternal
	case errors.Is(err, dao.ErrRecordNotFound):
		return ErrRecordNotFound
	case errors.Is(err, cache.ErrRecordNotFound):
		return ErrRecordNotFound
	case errors.Is(err, cache.ErrInvalidTokenData):
		return ErrInvalidToken
	case errors.Is(err, dao.ErrUniqueKey):
		return ErrUniqueKey
	default:
		return ErrServerInternal
	}
}
