package manager

import (
	"context"
	"log/slog"
	"time"

	search_grpc "github.com/yzletter/go-postery/api/proto/search/v1"
	"github.com/yzletter/go-postery/backend/grpc/errs"
)

type SearchServiceManager struct {
	service string
	hub     ServiceHub
}

func NewSearchManager(service string, hub ServiceHub) *SearchServiceManager {
	return &SearchServiceManager{
		service: service,
		hub:     hub,
	}
}

func (manager *SearchServiceManager) Search(ctx context.Context, req *search_grpc.SearchRequest) (*search_grpc.SearchResult, error) {
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
		client := search_grpc.NewSearchServiceClient(conn) // 建立 Client

		// 添加超时控制
		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *search_grpc.SearchResult
		resp, err = client.Search(ctx, req) // 微服务调用
		cancel()

		if isEndpointFailure(err) {
			endpoint.MarkFailed()
			slog.Error("call search service failed", "service", manager.service, "endpoint", endpoint.Addr, "try", try+1, "error", err)
			continue
		}
		endpoint.MarkSuccess()
		return resp, err // 返回 grpc 错误
	}

	// 默认会返回服务调用失败
	return nil, err
}

func (manager *SearchServiceManager) DeleteDoc(ctx context.Context, req *search_grpc.DocID) (*search_grpc.AffectedCount, error) {
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
		client := search_grpc.NewSearchServiceClient(conn) // 建立 Client

		// 添加超时控制
		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *search_grpc.AffectedCount
		resp, err = client.DeleteDoc(ctx, req) // 微服务调用
		cancel()

		if isEndpointFailure(err) {
			endpoint.MarkFailed()
			slog.Error("delete search document failed", "service", manager.service, "endpoint", endpoint.Addr, "doc_id", req.DocID, "try", try+1, "error", err)
			continue
		}
		endpoint.MarkSuccess()
		return resp, err // 返回 grpc 错误
	}

	// 默认会返回服务调用失败
	return nil, err
}

func (manager *SearchServiceManager) AddDoc(ctx context.Context, req *search_grpc.Document) (*search_grpc.AffectedCount, error) {
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
		client := search_grpc.NewSearchServiceClient(conn) // 建立 Client

		// 添加超时控制
		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *search_grpc.AffectedCount
		resp, err = client.AddDoc(ctx, req) // 微服务调用
		cancel()

		if isEndpointFailure(err) {
			endpoint.MarkFailed()
			slog.Error("add search document failed", "service", manager.service, "endpoint", endpoint.Addr, "doc_id", req.DocID, "try", try+1, "error", err)
			continue
		}
		endpoint.MarkSuccess()
		return resp, err // 返回 grpc 错误
	}

	// 默认会返回服务调用失败
	return nil, err
}

func (manager *SearchServiceManager) Count(ctx context.Context, req *search_grpc.CountRequest) (*search_grpc.AffectedCount, error) {
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
		client := search_grpc.NewSearchServiceClient(conn) // 建立 Client

		// 添加超时控制
		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *search_grpc.AffectedCount
		resp, err = client.Count(ctx, req) // 微服务调用
		cancel()

		if isEndpointFailure(err) {
			endpoint.MarkFailed()
			slog.Error("count search document failed", "service", manager.service, "endpoint", endpoint.Addr, "try", try+1, "error", err)
			continue
		}
		endpoint.MarkSuccess()
		return resp, err // 返回 grpc 错误
	}

	// 默认会返回服务调用失败
	return nil, err
}

func (manager *SearchServiceManager) StartHealthCheck(ctx context.Context) {
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

func (manager *SearchServiceManager) checkOnce(ctx context.Context) {
	endpoints := manager.hub.GetEndpoints(ctx, manager.service)
	for _, endpoint := range endpoints {
		if endpoint == nil {
			continue
		}
		conn := endpoint.ClientConn()
		if conn == nil {
			continue
		}

		client := search_grpc.NewSearchServiceClient(conn) // 建立 Client

		// 添加超时控制
		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		_, err := client.HealthCheck(ctx, &search_grpc.HealthCheckRequest{}) // 健康探测
		cancel()

		if err != nil {
			endpoint.MarkFailed()
			continue
		}
		endpoint.MarkSuccess()
	}
}
