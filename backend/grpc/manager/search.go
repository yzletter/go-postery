package manager

import (
	"context"
	"log/slog"
	"time"

	search_grpc "github.com/yzletter/go-postery/api/proto/search/v1"
	"github.com/yzletter/go-postery/backend/errs"
	search_model "github.com/yzletter/go-postery/microservice-backend/search/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	var err = errs.ErrUnavailable
	var tryCnt = 3
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil || endpoint.Conn == nil {
			continue
		}
		client := search_grpc.NewSearchServiceClient(endpoint.Conn)

		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *search_grpc.SearchResult
		resp, err = client.Search(ctx, req)
		cancel()

		if err != nil && status.Code(err) == codes.Internal {
			endpoint.MarkFailed()
			slog.Error("gRPC Error", "error", err, "service", manager.service, "endpoint", endpoint.Addr)
			continue
		}
		endpoint.MarkSuccess()
		return resp, err
	}

	return nil, err
}

func (manager *SearchServiceManager) DeleteDoc(ctx context.Context, req *search_grpc.DocID) (*search_grpc.AffectedCount, error) {
	var err = errs.ErrUnavailable
	var tryCnt = 1
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil || endpoint.Conn == nil {
			continue
		}
		client := search_grpc.NewSearchServiceClient(endpoint.Conn)

		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *search_grpc.AffectedCount
		resp, err = client.DeleteDoc(ctx, req)
		cancel()

		if err != nil && status.Code(err) == codes.Internal {
			endpoint.MarkFailed()
			slog.Error("gRPC Error", "error", err, "service", manager.service, "endpoint", endpoint.Addr)
			continue
		}
		endpoint.MarkSuccess()
		return resp, err
	}

	return nil, err
}

func (manager *SearchServiceManager) AddDoc(ctx context.Context, req *search_model.Document) (*search_grpc.AffectedCount, error) {
	var err = errs.ErrUnavailable
	var tryCnt = 1
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil || endpoint.Conn == nil {
			continue
		}
		client := search_grpc.NewSearchServiceClient(endpoint.Conn)

		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *search_grpc.AffectedCount
		resp, err = client.AddDoc(ctx, req)
		cancel()

		if err != nil && status.Code(err) == codes.Internal {
			endpoint.MarkFailed()
			slog.Error("gRPC Error", "error", err, "service", manager.service, "endpoint", endpoint.Addr)
			continue
		}
		endpoint.MarkSuccess()
		return resp, err
	}

	return nil, err
}

func (manager *SearchServiceManager) Count(ctx context.Context, req *search_grpc.CountRequest) (*search_grpc.AffectedCount, error) {
	var err = errs.ErrUnavailable
	var tryCnt = 3
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil || endpoint.Conn == nil {
			continue
		}
		client := search_grpc.NewSearchServiceClient(endpoint.Conn)

		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *search_grpc.AffectedCount
		resp, err = client.Count(ctx, req)
		cancel()

		if err != nil && status.Code(err) == codes.Internal {
			endpoint.MarkFailed()
			slog.Error("gRPC Error", "error", err, "service", manager.service, "endpoint", endpoint.Addr)
			continue
		}
		endpoint.MarkSuccess()
		return resp, err
	}

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
		if endpoint == nil || endpoint.Conn == nil {
			continue
		}

		client := search_grpc.NewSearchServiceClient(endpoint.Conn)
		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		_, err := client.HealthCheck(ctx, &search_grpc.HealthCheckRequest{})
		cancel()

		if err != nil {
			endpoint.MarkFailed()
			continue
		}
		endpoint.MarkSuccess()
	}
}
