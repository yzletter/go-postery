package server

import (
	"context"
	"log/slog"

	"github.com/yzletter/go-postery/microservice-backend/user/errs"
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
	method := info.FullMethod

	// 是否限流
	limited, err := interceptor.limiter.Limit(ctx, interceptor.limitPrefix, method)
	if err != nil {
		// 出错默认限流
		slog.Error("GrpcLimitInterceptor Limit Failed", "error", err)
		return nil, errs.ErrResourceExhausted
	}

	// 限流
	if limited {
		return nil, errs.ErrResourceExhausted
	}

	// 不限流
	return handler(ctx, req)
}
