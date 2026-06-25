package cache

import "errors"

var (
	// ErrReduceInventory 库存不足
	ErrReduceInventory = errors.New("库存已小于 0")
	// ErrServerInternal 缓存内部错误
	ErrServerInternal = errors.New("缓存内部错误")
)
