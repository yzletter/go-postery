package utils

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/yzletter/go-postery/errno"
)

func GetUidFromCTX(ctx *gin.Context, key string) (int64, error) {
	v, ok := ctx.Get(key)
	if !ok {
		return 0, errno.ErrUserNotLogin
	}
	uid, ok := v.(int64)
	if !ok {
		return 0, errno.ErrUserNotLogin
	}
	slog.Info("Get Uid From CTX Success", "uid", uid)
	return uid, nil
}
