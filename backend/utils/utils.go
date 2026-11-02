package utils

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

func GetUidFromCTX(ctx *gin.Context, key string) int64 {
	v, ok := ctx.Get(key)
	if !ok {
		return 0
	}
	uid, ok := v.(int64)
	if !ok {
		return 0
	}
	slog.Info("Get Uid From CTX Success", "uid", uid)
	return uid
}
