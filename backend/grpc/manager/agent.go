package manager

import (
	"context"
	"log/slog"
	"time"

	agent_grpc "github.com/yzletter/go-postery/api/proto/agent/v1"
	"github.com/yzletter/go-postery/backend/grpc/errs"
)

type AgentServiceManager struct {
	service string
	hub     ServiceHub
}

func NewAgentManager(service string, hub ServiceHub) *AgentServiceManager {
	return &AgentServiceManager{
		service: service,
		hub:     hub,
	}
}

func (manager *AgentServiceManager) Chat(ctx context.Context, req *agent_grpc.ChatRequest) (*agent_grpc.ChatResponse, error) {
	var err = errs.ErrUnavailable // 暴露错误
	var tryCnt = 1                // Chat 可能触发外部模型调用, 不重复提交
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil {
			continue
		}
		conn := endpoint.ClientConn()
		if conn == nil {
			continue
		}
		client := agent_grpc.NewAgentServiceClient(conn) // 建立 Client

		// 添加超时控制
		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *agent_grpc.ChatResponse
		resp, err = client.Chat(ctx, req) // 微服务调用
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

func (manager *AgentServiceManager) StartHealthCheck(ctx context.Context) {
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

func (manager *AgentServiceManager) checkOnce(ctx context.Context) {
	endpoints := manager.hub.GetEndpoints(ctx, manager.service)
	for _, endpoint := range endpoints {
		if endpoint == nil {
			continue
		}
		conn := endpoint.ClientConn()
		if conn == nil {
			continue
		}

		client := agent_grpc.NewAgentServiceClient(conn) // 建立 Client

		// 添加超时控制
		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		_, err := client.HealthCheck(ctx, &agent_grpc.HealthCheckRequest{}) // 健康探测
		cancel()

		if err != nil {
			endpoint.MarkFailed()
			continue
		}
		endpoint.MarkSuccess()
	}
}
