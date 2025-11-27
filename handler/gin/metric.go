package handler

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Counter是一个积累量（单调增），跟历史值有关
	requestCounter = promauto.NewCounterVec(prometheus.CounterOpts{Name: "request_counter"}, []string{"service", "interface"}) //此处指定了2个Label
	// Gauge是每个记录是独立的
	requestTimer = promauto.NewGaugeVec(prometheus.GaugeOpts{Name: "request_timer"}, []string{"service", "interface"})
)

// MetricHandler 返回每个接口的调用次数和调用时间
func MetricHandler(ctx *gin.Context) {
	// 记录开始时间
	start := time.Now()

	// 执行后面的中间件
	ctx.Next()

	// 当前接口调用 url, 需要对路径进行处理
	path := mapURL(ctx)

	requestCounter.WithLabelValues("gopostery", path).Inc()                                        // 计数器 + 1 即可
	requestTimer.WithLabelValues("gopostery", path).Set(float64(time.Since(start).Milliseconds())) // 计时器记录从 start 到现在过了多久
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
