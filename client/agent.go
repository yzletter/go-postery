package client

import (
	"context"
	"time"

	agent_grpc "github.com/yzletter/go-postery/api/proto/agent/v1"
	"google.golang.org/grpc"
)

type agentClient struct {
	conn   *grpc.ClientConn
	client agent_grpc.AgentServiceClient
}

func NewAgentClient(conn *grpc.ClientConn) (AgentClient, error) {
	if err := validateConn(conn); err != nil {
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
