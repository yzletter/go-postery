package utils

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yzletter/go-postery/service"
)

var (
	ErrNotLogin = errors.New("请先登录")
)

func GetUidFromCTX(ctx *gin.Context) (int, error) {
	idStr, ok := ctx.Value(service.UID_IN_CTX).(string)
	if !ok {
		return 0, ErrNotLogin
	}

	uid, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		// 没有登录
		return 0, ErrNotLogin
	}

	return int(uid), nil
}
