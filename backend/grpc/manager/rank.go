package manager

import (
	"context"
	"log/slog"
	"time"

	rank_grpc "github.com/yzletter/go-postery/api/proto/rank/v1"
	"github.com/yzletter/go-postery/backend/grpc/errs"
)

type RankServiceManager struct {
	service string
	hub     ServiceHub
}

func NewRankManager(service string, hub ServiceHub) *RankServiceManager {
	return &RankServiceManager{
		service: service,
		hub:     hub,
	}
}

func (manager *RankServiceManager) RankUser(ctx context.Context, req *rank_grpc.RankIDRequest) (*rank_grpc.RankEmptyResponse, error) {
	var err = errs.ErrUnavailable // 暴露错误
	var tryCnt = 1                // 写入类调用只适合一次
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil {
			continue
		}
		conn := endpoint.ClientConn()
		if conn == nil {
			continue
		}
		client := rank_grpc.NewRankServiceClient(conn) // 建立 Client

		// 添加超时控制
		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *rank_grpc.RankEmptyResponse
		resp, err = client.RankUser(ctx, req) // 微服务调用
		cancel()

		if isEndpointFailure(err) {
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

func (manager *RankServiceManager) RankPost(ctx context.Context, req *rank_grpc.RankIDRequest) (*rank_grpc.RankEmptyResponse, error) {
	var err = errs.ErrUnavailable // 暴露错误
	var tryCnt = 1                // 写入类调用只适合一次
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil {
			continue
		}
		conn := endpoint.ClientConn()
		if conn == nil {
			continue
		}
		client := rank_grpc.NewRankServiceClient(conn) // 建立 Client

		// 添加超时控制
		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *rank_grpc.RankEmptyResponse
		resp, err = client.RankPost(ctx, req) // 微服务调用
		cancel()

		if isEndpointFailure(err) {
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

func (manager *RankServiceManager) RankTopKUser(ctx context.Context, req *rank_grpc.RankEmptyRequest) (*rank_grpc.RankEmptyResponse, error) {
	var err = errs.ErrUnavailable // 暴露错误
	var tryCnt = 1                // 写入类调用只适合一次
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil {
			continue
		}
		conn := endpoint.ClientConn()
		if conn == nil {
			continue
		}
		client := rank_grpc.NewRankServiceClient(conn) // 建立 Client

		// 添加超时控制
		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *rank_grpc.RankEmptyResponse
		resp, err = client.RankTopKUser(ctx, req) // 微服务调用
		cancel()

		if isEndpointFailure(err) {
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

func (manager *RankServiceManager) RankTopKPost(ctx context.Context, req *rank_grpc.RankEmptyRequest) (*rank_grpc.RankEmptyResponse, error) {
	var err = errs.ErrUnavailable // 暴露错误
	var tryCnt = 1                // 写入类调用只适合一次
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil {
			continue
		}
		conn := endpoint.ClientConn()
		if conn == nil {
			continue
		}
		client := rank_grpc.NewRankServiceClient(conn) // 建立 Client

		// 添加超时控制
		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *rank_grpc.RankEmptyResponse
		resp, err = client.RankTopKPost(ctx, req) // 微服务调用
		cancel()

		if isEndpointFailure(err) {
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

func (manager *RankServiceManager) TopKUser(ctx context.Context, req *rank_grpc.RankEmptyRequest) (*rank_grpc.TopKUserResponse, error) {
	var err = errs.ErrUnavailable // 暴露错误
	var tryCnt = 3                // 查询类调用可重试
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil {
			continue
		}
		conn := endpoint.ClientConn()
		if conn == nil {
			continue
		}
		client := rank_grpc.NewRankServiceClient(conn) // 建立 Client

		// 添加超时控制
		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *rank_grpc.TopKUserResponse
		resp, err = client.TopKUser(ctx, req) // 微服务调用
		cancel()

		if isEndpointFailure(err) {
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

func (manager *RankServiceManager) TopKPost(ctx context.Context, req *rank_grpc.RankEmptyRequest) (*rank_grpc.TopKPostResponse, error) {
	var err = errs.ErrUnavailable // 暴露错误
	var tryCnt = 3                // 查询类调用可重试
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil {
			continue
		}
		conn := endpoint.ClientConn()
		if conn == nil {
			continue
		}
		client := rank_grpc.NewRankServiceClient(conn) // 建立 Client

		// 添加超时控制
		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *rank_grpc.TopKPostResponse
		resp, err = client.TopKPost(ctx, req) // 微服务调用
		cancel()

		if isEndpointFailure(err) {
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

func (manager *RankServiceManager) StartHealthCheck(ctx context.Context) {
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

func (manager *RankServiceManager) checkOnce(ctx context.Context) {
	endpoints := manager.hub.GetEndpoints(ctx, manager.service)
	for _, endpoint := range endpoints {
		if endpoint == nil {
			continue
		}
		conn := endpoint.ClientConn()
		if conn == nil {
			continue
		}

		client := rank_grpc.NewRankServiceClient(conn) // 建立 Client

		// 添加超时控制
		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		_, err := client.HealthCheck(ctx, &rank_grpc.HealthCheckRequest{}) // 健康探测
		cancel()

		if err != nil {
			endpoint.MarkFailed()
			continue
		}
		endpoint.MarkSuccess()
	}
}
