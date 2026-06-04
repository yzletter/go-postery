package cache

import "errors"

var (
	ErrReduceInventory   = errors.New("库存已小于 0")
	ErrRollbackInventory = errors.New("库存回补失败")
	ErrCreateTempOrder   = errors.New("临时订单创建失败")
	ErrTempOrderMissing  = errors.New("临时订单不存在")
	ErrRecycleTempOrder  = errors.New("临时订单回收失败")
)
