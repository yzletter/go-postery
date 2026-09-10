package manager

import (
	"context"
	"log/slog"
	"time"

	oss_grpc2 "github.com/yzletter/go-postery/backend/api/proto/oss/v1"
	"github.com/yzletter/go-postery/backend/grpc/errs"
)

type OSSServiceManager struct {
	service string
	hub     ServiceHub
}

func NewOSSManager(ctx context.Context, service string, hub ServiceHub) *OSSServiceManager {
	hub.LoadEndpoints(ctx, service)
	hub.WatchEndpointsFromServiceHub(ctx, service)

	manager := &OSSServiceManager{service: service, hub: hub}
	go manager.startHealthCheck(ctx) // 开启下游服务健康检查

	return manager
}

func (manager *OSSServiceManager) SignUpload(ctx context.Context, req *oss_grpc2.SignUploadRequest) (*oss_grpc2.SignUploadResponse, error) {
	var err = errs.ErrUnavailable
	var tryCnt = 3
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil {
			continue
		}
		conn := endpoint.ClientConn()
		if conn == nil {
			continue
		}
		client := oss_grpc2.NewOSSServiceClient(conn)

		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *oss_grpc2.SignUploadResponse
		resp, err = client.SignUpload(ctx, req)
		cancel()

		if isEndpointFailure(err) {
			endpoint.MarkFailed()
			slog.Error("gRPC Error", "error", err, "service", manager.service, "endpoint", endpoint.Addr)
			continue
		}
		endpoint.MarkSuccess()
		return resp, err
	}

	return nil, err
}

func (manager *OSSServiceManager) UploadCallback(ctx context.Context, req *oss_grpc2.UploadCallbackRequest) (*oss_grpc2.UploadCallbackResponse, error) {
	var err = errs.ErrUnavailable
	var tryCnt = 1
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil {
			continue
		}
		conn := endpoint.ClientConn()
		if conn == nil {
			continue
		}
		client := oss_grpc2.NewOSSServiceClient(conn)

		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *oss_grpc2.UploadCallbackResponse
		resp, err = client.UploadCallback(ctx, req)
		cancel()

		if isEndpointFailure(err) {
			endpoint.MarkFailed()
			slog.Error("gRPC Error", "error", err, "service", manager.service, "endpoint", endpoint.Addr)
			continue
		}
		endpoint.MarkSuccess()
		return resp, err
	}

	return nil, err
}

func (manager *OSSServiceManager) GetObjectURL(ctx context.Context, req *oss_grpc2.GetObjectURLRequest) (*oss_grpc2.GetObjectURLResponse, error) {
	var err = errs.ErrUnavailable
	var tryCnt = 3
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil {
			continue
		}
		conn := endpoint.ClientConn()
		if conn == nil {
			continue
		}
		client := oss_grpc2.NewOSSServiceClient(conn)

		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *oss_grpc2.GetObjectURLResponse
		resp, err = client.GetObjectURL(ctx, req)
		cancel()

		if isEndpointFailure(err) {
			endpoint.MarkFailed()
			slog.Error("gRPC Error", "error", err, "service", manager.service, "endpoint", endpoint.Addr)
			continue
		}
		endpoint.MarkSuccess()
		return resp, err
	}

	return nil, err
}

func (manager *OSSServiceManager) startHealthCheck(ctx context.Context) {
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

func (manager *OSSServiceManager) checkOnce(ctx context.Context) {
	endpoints := manager.hub.GetEndpoints(ctx, manager.service)
	for _, endpoint := range endpoints {
		if endpoint == nil {
			continue
		}
		conn := endpoint.ClientConn()
		if conn == nil {
			continue
		}

		client := oss_grpc2.NewOSSServiceClient(conn)

		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		_, err := client.HealthCheck(ctx, &oss_grpc2.HealthCheckRequest{})
		cancel()

		if err != nil {
			endpoint.MarkFailed()
			continue
		}

		endpoint.MarkSuccess()
	}
}
