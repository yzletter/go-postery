package dao

import "errors"

var (
	// ErrServerInternal 数据库内部错误
	ErrServerInternal = errors.New("数据库内部错误")
	// ErrRecordNotFound 记录不存在
	ErrRecordNotFound = errors.New("记录不存在")
	// ErrUniqueKey 唯一键冲突
	ErrUniqueKey = errors.New("唯一键冲突")
	// ErrParamsInvalid 参数错误
	ErrParamsInvalid = errors.New("参数有误")
)
