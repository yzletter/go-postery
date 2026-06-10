package manager

import (
	"context"
	"log/slog"
	"time"

	code_grpc "github.com/yzletter/go-postery/api/proto/code/v1"
	"github.com/yzletter/go-postery/backend/errs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CodeServiceManager struct {
	service string
	hub     ServiceHub
}

func NewCodeManager(service string, hub ServiceHub) *CodeServiceManager {
	return &CodeServiceManager{
		service: service,
		hub:     hub,
	}
}

func (manager *CodeServiceManager) Send(ctx context.Context, req *code_grpc.SendCodeRequest) (*code_grpc.SendCodeResponse, error) {
	var err = errs.ErrUnavailable // 暴露错误
	var tryCnt = 1                // Send 只适合一次
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil || endpoint.Conn == nil {
			continue
		}
		client := code_grpc.NewCodeServiceClient(endpoint.Conn) // 建立 Client

		// 添加超时控制
		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *code_grpc.SendCodeResponse
		resp, err = client.Send(ctx, req) // 微服务调用
		cancel()

		if err != nil && status.Code(err) == codes.Internal {
			endpoint.MarkFailed()
			slog.Error("gRPC Error", "error", err, "service", manager.service, "endpoint", endpoint.Addr)
			continue
		}
		endpoint.MarkSuccess()
		return resp, err // 返回 grpc 错误
	}

	// 默认会返回服务调用失败
	return nil, err
}

func (manager *CodeServiceManager) Verify(ctx context.Context, req *code_grpc.CheckCodeRequest) (*code_grpc.CheckCodeResponse, error) {
	var err = errs.ErrUnavailable // 暴露错误
	var tryCnt = 1                // Verify 只适合一次
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil || endpoint.Conn == nil {
			continue
		}
		client := code_grpc.NewCodeServiceClient(endpoint.Conn) // 建立 Client

		// 添加超时控制
		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *code_grpc.CheckCodeResponse
		resp, err = client.Verify(ctx, req) // 微服务调用
		cancel()

		if err != nil && status.Code(err) == codes.Internal {
			endpoint.MarkFailed()
			slog.Error("gRPC Error", "error", err, "service", manager.service, "endpoint", endpoint.Addr)
			continue
		}
		endpoint.MarkSuccess()
		return resp, err // 返回 grpc 错误
	}

	// 默认会返回服务调用失败
	return nil, err
}

func (manager *CodeServiceManager) StartHealthCheck(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			manager.checkOnce(ctx)
		}
	}
}

func (manager *CodeServiceManager) checkOnce(ctx context.Context) {
	endpoints := manager.hub.GetEndpoints(ctx, manager.service)
	for _, endpoint := range endpoints {
		if endpoint == nil || endpoint.Conn == nil {
			continue
		}

		client := code_grpc.NewCodeServiceClient(endpoint.Conn) // 建立 Client

		// 添加超时控制
		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		_, err := client.HealthCheck(ctx, &code_grpc.HealthCheckRequest{}) // 健康探测
		cancel()

		if err != nil {
			endpoint.MarkFailed()
			continue
		}

		endpoint.MarkSuccess()
	}
}
