package dao

import "errors"

// 定义 DAO 层所有错误

var (
	ErrServerInternal = errors.New("数据库内部错误")
	ErrRecordNotFound = errors.New("记录不存在")
	ErrUniqueKey      = errors.New("唯一键冲突")
	ErrParamsInvalid  = errors.New("参数有误")
)
