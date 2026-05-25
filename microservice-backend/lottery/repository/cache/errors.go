package cache

import "errors"

var (
	ErrReduceInventory = errors.New("库存已小于 0")
	ErrCreateTempOrder = errors.New("临时订单创建失败")
)
