package cache

import "errors"

var (
	ErrReduceInventory   = errors.New("库存已小于 0")
	ErrRollbackInventory = errors.New("库存回补失败")
)
