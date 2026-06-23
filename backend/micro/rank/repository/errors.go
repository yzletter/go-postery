package repository

import "errors"

var (
	ErrServerInternal   = errors.New("内部错误")
	ErrRecordNotFound   = errors.New("资源不存在")
	ErrUniqueKey        = errors.New("唯一键冲突")
	ErrResourceConflict = errors.New("资源冲突")
)

func toRepositoryErr(err error) error {
	return ErrServerInternal
}
