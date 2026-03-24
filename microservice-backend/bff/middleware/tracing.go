package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// TracingMiddleware 为 Gin 请求创建服务端 span，并把 span context 写回 Request.Context。
func TracingMiddleware(service string) gin.HandlerFunc {
	tracer := otel.Tracer(service)

	return func(ctx *gin.Context) {
		reqCtx := otel.GetTextMapPropagator().Extract(ctx.Request.Context(), propagation.HeaderCarrier(ctx.Request.Header))
		route := ctx.FullPath()
		if route == "" {
			route = ctx.Request.URL.Path
		}

		reqCtx, span := tracer.Start(reqCtx, spanName(ctx.Request.Method, route), trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()

		ctx.Request = ctx.Request.WithContext(reqCtx)

		span.SetAttributes(
			attribute.String("http.request.method", ctx.Request.Method),
			attribute.String("url.path", ctx.Request.URL.Path),
			attribute.String("url.query", ctx.Request.URL.RawQuery),
		)
		if route != "" {
			span.SetAttributes(attribute.String("http.route", route))
		}

		ctx.Next()

		if fullPath := ctx.FullPath(); fullPath != "" {
			span.SetName(spanName(ctx.Request.Method, fullPath))
			span.SetAttributes(attribute.String("http.route", fullPath))
		}

		statusCode := ctx.Writer.Status()
		if statusCode == 0 {
			statusCode = http.StatusOK
		}
		span.SetAttributes(attribute.Int("http.response.status_code", statusCode))

		if len(ctx.Errors) > 0 {
			err := ctx.Errors.Last()
			span.RecordError(err)
			span.SetStatus(otelcodes.Error, err.Error())
			return
		}
		if statusCode >= http.StatusInternalServerError {
			span.SetStatus(otelcodes.Error, http.StatusText(statusCode))
			return
		}

		span.SetStatus(otelcodes.Ok, http.StatusText(statusCode))
	}
}

func spanName(method, route string) string {
	if route == "" {
		return method
	}
	return method + " " + route
}
