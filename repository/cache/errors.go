package cache

import "errors"

var (
	ErrReduceInventory = errors.New("库存已小于 0")
)
