package cache

import "errors"

var (
	// ErrRecordNotFound 缓存记录不存在
	ErrRecordNotFound = errors.New("record not found")
	// ErrInvalidTokenData 缓存中的 Token 数据不完整或不合法
	ErrInvalidTokenData = errors.New("invalid token data")
)
