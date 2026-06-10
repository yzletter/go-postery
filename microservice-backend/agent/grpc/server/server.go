package server

import (
	"context"

	agent_grpc "github.com/yzletter/go-postery/api/proto/agent/v1"
	"github.com/yzletter/go-postery/microservice-backend/agent/service"
)

type AgentServiceServer struct {
	svc service.AgentService
	agent_grpc.UnimplementedAgentServiceServer
}

func NewAgentServiceServer(svc service.AgentService) *AgentServiceServer {
	return &AgentServiceServer{
		svc: svc,
	}
}

func (server *AgentServiceServer) Chat(ctx context.Context, req *agent_grpc.ChatRequest) (*agent_grpc.ChatResponse, error) {
	content, documents, err := server.svc.Chat(ctx, req.UserID, req.SessionID, req.Query)
	if err != nil {
		return &agent_grpc.ChatResponse{}, err
	}

	return &agent_grpc.ChatResponse{
		SessionID: req.SessionID,
		Content:   content,
		Documents: documents,
	}, nil
}

func (server *AgentServiceServer) HealthCheck(ctx context.Context, req *agent_grpc.HealthCheckRequest) (*agent_grpc.HealthCheckResponse, error) {
	return &agent_grpc.HealthCheckResponse{}, nil
}
