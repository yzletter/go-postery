package repository

import (
	"errors"

	"github.com/yzletter/go-postery/backend/micro/lottery/repository/cache"
	"github.com/yzletter/go-postery/backend/micro/lottery/repository/dao"
)

var (
	ErrInsufficientStock = errors.New("库存不足")
	ErrServerInternal    = errors.New("系统内部错误")
	ErrRecordNotFound    = errors.New("资源不存在")
	ErrUniqueKey         = errors.New("唯一键冲突")
	ErrResourceConflict  = errors.New("资源请求失败")
)

func toRepositoryErr(err error) error {
	switch {
	case errors.Is(err, dao.ErrServerInternal):
		return ErrServerInternal
	case errors.Is(err, dao.ErrRecordNotFound):
		return ErrRecordNotFound
	case errors.Is(err, cache.ErrRecordNotFound):
		return ErrRecordNotFound
	case errors.Is(err, cache.ErrReduceInventory):
		return ErrInsufficientStock
	case errors.Is(err, dao.ErrUniqueKey):
		return ErrUniqueKey
	default:
		return ErrServerInternal
	}
}
