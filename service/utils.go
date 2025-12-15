package service

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

var (
	ErrNotLogin = errors.New("请先登录")
)

func GetUidFromCTX(ctx *gin.Context) (int64, error) {
	idStr, ok := ctx.Value(UID_IN_CTX).(string)
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
