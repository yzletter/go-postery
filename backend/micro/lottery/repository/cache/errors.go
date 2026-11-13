package cache

import "errors"

var (
	ErrReduceInventory   = errors.New("库存已小于 0")
	ErrInvalidInventory  = errors.New("库存数据异常")
	ErrRollbackInventory = errors.New("库存回补失败")
	ErrInvalidParameter  = errors.New("非法数据")
	ErrRecordNotFound    = errors.New("记录不存在")
)
