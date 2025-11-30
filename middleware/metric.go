package middleware

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yzletter/go-postery/handler"
)

// MetricMiddleware 返回每个接口的调用次数和调用时间
func MetricMiddleware(metricHandler *handler.MetricHandler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 记录开始时间
		start := time.Now()

		// 执行后面的中间件
		ctx.Next()

		// 当前接口调用 url, 需要对路径进行处理
		path := mapURL(ctx)

		// 对该路径的请求进行统计
		metricHandler.CounterAdd(path)      // 计数器 +1
		metricHandler.TimerSet(path, start) // 计时器记录时间

	}
}

var mpRestful = map[string]string{"id": "id"}

func mapURL(ctx *gin.Context) string {
	url := ctx.Request.URL.Path
	// ctx.Params 返回请求参数切片, 切片中每个元素为 {Key, Val}
	for _, param := range ctx.Params {
		if value, ok := mpRestful[param.Key]; ok != false {
			url = strings.Replace(url, param.Value, value, 1) // 把具体值换成抽象 eg : /posts/delete/3 -> /posts/delete/:id
		}
	}
	return url
}
