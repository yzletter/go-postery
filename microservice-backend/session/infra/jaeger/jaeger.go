package jaeger

import (
	"context"
	"log"
	"time"

	"github.com/yzletter/go-postery/microservice-backend/session/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.39.0"
)

// InitJaeger 初始化 Jaeger
func InitJaeger(ctx context.Context, config config.JaegerConfig, service string) func() {
	// 构造 Jaeger Trace Provider
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(config.Addr),
		otlptracehttp.WithInsecure())
	if err != nil {
		return nil
	}

	res, err := resource.New(ctx, resource.WithAttributes(semconv.ServiceName(service)))
	if err != nil {
		return nil
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()), // 采样
		sdktrace.WithBatcher(exporter, sdktrace.WithBatchTimeout(time.Second)),
	)

	// 设置 Trace Provider
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	// 返回 Shutdown
	return func() {
		if err := provider.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}
}
