package repository

import (
	"errors"

	"github.com/yzletter/go-postery/backend/micro/user/repository/dao"
)

var (
	// ErrServerInternal 系统内部错误
	ErrServerInternal = errors.New("内部错误")
	// ErrRecordNotFound 资源不存在
	ErrRecordNotFound = errors.New("资源不存在")
	// ErrUniqueKey 唯一键冲突
	ErrUniqueKey = errors.New("唯一键冲突")
	// ErrResourceConflict 资源冲突
	ErrResourceConflict = errors.New("资源冲突")
)

// toRepositoryErr 将 DAO 错误转换为 Repository 错误
func toRepositoryErr(err error) error {
	switch {
	case errors.Is(err, dao.ErrServerInternal):
		return ErrServerInternal
	case errors.Is(err, dao.ErrRecordNotFound):
		return ErrRecordNotFound
	case errors.Is(err, dao.ErrUniqueKey):
		return ErrUniqueKey
	default:
		return ErrServerInternal
	}
}
