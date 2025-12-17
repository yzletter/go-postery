package utils

import (
	"context"
	"errors"
	"strconv"
)

var (
	ErrNotLogin = errors.New("请先登录")
)

func GetUidFromCTX(ctx context.Context, key string) (int64, error) {
	idStr, ok := ctx.Value(key).(string)
	if !ok {
		return 0, ErrNotLogin
	}

	uid, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		// 没有登录
		return 0, ErrNotLogin
	}

	return uid, nil
}
