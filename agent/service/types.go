package service

import (
	"context"

	agent_grpc "github.com/yzletter/go-postery/api/proto/agent/v1"
)

type AgentService interface {
	Chat(context.Context, *agent_grpc.ChatRequest) (*agent_grpc.ChatResponse, error)
	StartChunkDocConsumer(ctx context.Context)
	StartUpsertQdrantConsumer(ctx context.Context)
	agent_grpc.UnsafeAgentServiceServer
}
