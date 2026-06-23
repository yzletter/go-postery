package service

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type MetricService struct {
	requestCounter *prometheus.CounterVec
	requestTimer   *prometheus.GaugeVec
}

func NewMetricService() *MetricService {
	return &MetricService{
		requestCounter: promauto.NewCounterVec(prometheus.CounterOpts{Name: "request_counter"}, []string{"service", "interface"}),
		requestTimer:   promauto.NewGaugeVec(prometheus.GaugeOpts{Name: "request_timer"}, []string{"service", "interface"}),
	}
}

func (svc *MetricService) CounterAdd(path string) {
	svc.requestCounter.WithLabelValues("gopostery", path).Inc()
}

func (svc *MetricService) TimerSet(path string, start time.Time) {
	svc.requestTimer.WithLabelValues("gopostery", path).Set(float64(time.Since(start).Milliseconds()))
}
