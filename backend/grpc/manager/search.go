package manager

import (
	"context"
	"log/slog"
	"time"

	search_grpc2 "github.com/yzletter/go-postery/backend/api/proto/search/v1"
	"github.com/yzletter/go-postery/backend/grpc/errs"
)

type SearchServiceManager struct {
	service string
	hub     ServiceHub
}

func NewSearchManager(ctx context.Context, service string, hub ServiceHub) *SearchServiceManager {
	hub.LoadEndpoints(ctx, service)
	hub.WatchEndpointsFromServiceHub(ctx, service)

	manager := &SearchServiceManager{service: service, hub: hub}
	go manager.startHealthCheck(ctx) // 开启下游服务健康检查

	return manager
}

func (manager *SearchServiceManager) Search(ctx context.Context, req *search_grpc2.SearchRequest) (*search_grpc2.SearchResult, error) {
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
		client := search_grpc2.NewSearchServiceClient(conn) // 建立 Client

		// 添加超时控制
		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *search_grpc2.SearchResult
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

func (manager *SearchServiceManager) DeleteDoc(ctx context.Context, req *search_grpc2.DocID) (*search_grpc2.AffectedCount, error) {
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
		client := search_grpc2.NewSearchServiceClient(conn) // 建立 Client

		// 添加超时控制
		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *search_grpc2.AffectedCount
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

func (manager *SearchServiceManager) AddDoc(ctx context.Context, req *search_grpc2.Document) (*search_grpc2.AffectedCount, error) {
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
		client := search_grpc2.NewSearchServiceClient(conn) // 建立 Client

		// 添加超时控制
		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *search_grpc2.AffectedCount
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

func (manager *SearchServiceManager) Count(ctx context.Context, req *search_grpc2.CountRequest) (*search_grpc2.AffectedCount, error) {
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
		client := search_grpc2.NewSearchServiceClient(conn) // 建立 Client

		// 添加超时控制
		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *search_grpc2.AffectedCount
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

func (manager *SearchServiceManager) startHealthCheck(ctx context.Context) {
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

		client := search_grpc2.NewSearchServiceClient(conn) // 建立 Client

		// 添加超时控制
		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		_, err := client.HealthCheck(ctx, &search_grpc2.HealthCheckRequest{}) // 健康探测
		cancel()

		if err != nil {
			endpoint.MarkFailed()
			continue
		}
		endpoint.MarkSuccess()
	}
}
