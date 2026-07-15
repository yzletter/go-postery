package cache

import "errors"

var (
	ErrServerInternal = errors.New("缓存内部错误")
	ErrParamsInvalid  = errors.New("参数有误")
)
