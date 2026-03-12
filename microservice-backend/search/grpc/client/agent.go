package client

import (
	"context"
	"time"

	agent_grpc "github.com/yzletter/go-postery/api/proto/agent/v1"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

type agentClient struct {
	conn   *grpc.ClientConn
	client agent_grpc.AgentServiceClient
}

func NewAgentClient() (AgentClient, error) {
	// 建议：启用 ka，避免中间网络设备把长连接静默掐掉
	ka := keepalive.ClientParameters{
		Time:                30 * time.Second,
		Timeout:             10 * time.Second,
		PermitWithoutStream: true,
	}

	conn, err := grpc.NewClient(
		AgentClientAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()), // 生产用 TLS
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),       // Jaeger
		grpc.WithKeepaliveParams(ka),
	)
	if err != nil {
		return nil, err
	}

	return &agentClient{
		conn:   conn,
		client: agent_grpc.NewAgentServiceClient(conn),
	}, nil
}

func (client *agentClient) Close() {
	_ = client.conn.Close()
}

func (client *agentClient) Chat(ctx context.Context, req *agent_grpc.ChatRequest) (*agent_grpc.ChatResponse, error) {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(ctx, 8000*time.Millisecond)
	defer cancel()

	return client.client.Chat(ctx, req)
}
