package pkg

import (
	"context"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"google.golang.org/grpc"
)

type MetricService struct {
	service    string                 // 观测的服务名
	reqCounter *prometheus.CounterVec // 计数器
	reqTimer   *prometheus.GaugeVec   // 计时器
}

func NewMetricService(service string) *MetricService {
	return &MetricService{
		service: service,
		reqCounter: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: "go_postery",
			Subsystem: service,
			Name:      "request_counter",
		}, []string{"service", "interface"}),
		reqTimer: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "go_postery",
			Subsystem: service,
			Name:      "request_timer",
		}, []string{"service", "interface"}),
	}
}

func NewMetricServiceWithRegistry(service string, registry *prometheus.Registry) *MetricService {
	factory := promauto.With(registry)
	return &MetricService{
		service: service,
		reqCounter: factory.NewCounterVec(prometheus.CounterOpts{
			Namespace: "go_postery",
			Subsystem: service,
			Name:      "request_counter",
		}, []string{"service", "interface"}),
		reqTimer: factory.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "go_postery",
			Subsystem: service,
			Name:      "request_timer",
		}, []string{"service", "interface"}),
	}
}

// TimerInterceptor gRPC 计时拦截器
func (svc *MetricService) TimerInterceptor() func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		start := time.Now()
		resp, err = handler(ctx, req)
		svc.TimerSet(getMethod(info.FullMethod), start)
		return
	}
}

// CounterInterceptor gRPC 计数拦截器
func (svc *MetricService) CounterInterceptor() func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		svc.CounterAdd(getMethod(info.FullMethod))
		resp, err = handler(ctx, req)
		return
	}
}

// CounterAdd 计数
func (svc *MetricService) CounterAdd(key string) {
	svc.reqCounter.WithLabelValues(svc.service, key).Inc() // 计数器 + 1 即可
}

// TimerSet 计时
func (svc *MetricService) TimerSet(key string, start time.Time) {
	svc.reqTimer.WithLabelValues(svc.service, key).Set(float64(time.Since(start).Milliseconds())) // 计时器记录从 start 到现在过了多久
}

// 取 gRPC 调用的具体方法
func getMethod(fullMethod string) string {
	segments := strings.Split(fullMethod, "/")
	return segments[len(segments)-1]
}
