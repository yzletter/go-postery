package service

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type MetricService struct {
	requestCounter *prometheus.CounterVec // Counter 是一个积累量（单调增），跟历史值有关
	requestTimer   *prometheus.GaugeVec   // Gauge 是每个记录是独立的
}

func NewMetricService() *MetricService {
	return &MetricService{
		requestCounter: promauto.NewCounterVec(prometheus.CounterOpts{Name: "request_counter"}, []string{"service", "interface"}), //此处指定了2个Label,
		requestTimer:   promauto.NewGaugeVec(prometheus.GaugeOpts{Name: "request_timer"}, []string{"service", "interface"}),
	}
}

func (svc *MetricService) CounterAdd(path string) {
	svc.requestCounter.WithLabelValues("gopostery", path).Inc() // 计数器 + 1 即可
}

func (svc *MetricService) TimerSet(path string, start time.Time) {
	svc.requestTimer.WithLabelValues("gopostery", path).Set(float64(time.Since(start).Milliseconds())) // 计时器记录从 start 到现在过了多久

}
