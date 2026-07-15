package manager

import (
	"context"
	"log/slog"
	"time"

	ws_gateway_grpc "github.com/yzletter/go-postery/api/proto/ws_gateway/v1"
	"github.com/yzletter/go-postery/backend/grpc/errs"
)

type WSGatewayServiceManager struct {
	service string
	hub     ServiceHub
}

func NewWSGatewayManager(ctx context.Context, service string, hub ServiceHub) *WSGatewayServiceManager {
	hub.LoadEndpoints(ctx, service)
	hub.WatchEndpointsFromServiceHub(ctx, service)

	manager := &WSGatewayServiceManager{service: service, hub: hub}
	go manager.startHealthCheck(ctx) // 开启下游服务健康检查

	return manager
}

func (manager *WSGatewayServiceManager) Push(ctx context.Context, req *ws_gateway_grpc.PushRequest) (*ws_gateway_grpc.PushResponse, error) {
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
		client := ws_gateway_grpc.NewWSGatewayServiceClient(conn)

		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *ws_gateway_grpc.PushResponse
		resp, err = client.Push(ctx, req)
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

func (manager *WSGatewayServiceManager) startHealthCheck(ctx context.Context) {
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

func (manager *WSGatewayServiceManager) checkOnce(ctx context.Context) {
	endpoints := manager.hub.GetEndpoints(ctx, manager.service)
	for _, endpoint := range endpoints {
		if endpoint == nil {
			continue
		}
		conn := endpoint.ClientConn()
		if conn == nil {
			continue
		}

		client := ws_gateway_grpc.NewWSGatewayServiceClient(conn)

		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		_, err := client.HealthCheck(ctx, &ws_gateway_grpc.HealthCheckRequest{})
		cancel()

		if err != nil {
			endpoint.MarkFailed()
			continue
		}
		endpoint.MarkSuccess()
	}
}
