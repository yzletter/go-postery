package grpc

import (
	"context"
	"log/slog"
	"strings"

	"github.com/yzletter/go-postery/backend/grpc/errs"
	_grpc "google.golang.org/grpc"
)

type GrpcLimitInterceptor struct {
	limitPrefix string
	limiter     Limiter
}

type Limiter interface {
	Limit(ctx context.Context, prefix, identifier string) (bool, error)
}

func NewGrpcLimitInterceptor(limitPrefix string, limiter Limiter) *GrpcLimitInterceptor {
	return &GrpcLimitInterceptor{
		limitPrefix: limitPrefix,
		limiter:     limiter,
	}
}

// BuildLimiter 服务端限流拦截器
func (interceptor *GrpcLimitInterceptor) BuildLimiter(ctx context.Context, req any, info *_grpc.UnaryServerInfo, handler _grpc.UnaryHandler) (resp any, err error) {
	// HealthCheck 放行
	if strings.HasSuffix(info.FullMethod, "/HealthCheck") {
		return handler(ctx, req)
	}

	// 其他方法进行限流
	limited, err := interceptor.limiter.Limit(ctx, interceptor.limitPrefix, info.FullMethod)
	if err != nil {
		slog.Error("GrpcLimitInterceptor Limit Failed", "error", err)
		return nil, errs.ErrResourceExhausted
	}
	if limited {
		return nil, errs.ErrResourceExhausted
	}
	return handler(ctx, req)
}
