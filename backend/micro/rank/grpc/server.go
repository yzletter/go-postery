package grpc

import (
	"context"

	rank_grpc "github.com/yzletter/go-postery/api/proto/rank/v1"
	"github.com/yzletter/go-postery/backend/micro/rank/service"
)

type RankServiceServer struct {
	svc service.RankService
	rank_grpc.UnimplementedRankServiceServer
}

func NewRankServiceServer(svc service.RankService) *RankServiceServer {
	return &RankServiceServer{
		svc: svc,
	}
}

func (server *RankServiceServer) RankUser(ctx context.Context, req *rank_grpc.RankIDRequest) (*rank_grpc.RankEmptyResponse, error) {
	if err := server.svc.RankUser(ctx, req.ID); err != nil {
		return &rank_grpc.RankEmptyResponse{}, err
	}
	return &rank_grpc.RankEmptyResponse{}, nil
}

func (server *RankServiceServer) RankPost(ctx context.Context, req *rank_grpc.RankIDRequest) (*rank_grpc.RankEmptyResponse, error) {
	if err := server.svc.RankPost(ctx, req.ID); err != nil {
		return &rank_grpc.RankEmptyResponse{}, err
	}
	return &rank_grpc.RankEmptyResponse{}, nil
}

func (server *RankServiceServer) RankTopKUser(ctx context.Context, req *rank_grpc.RankEmptyRequest) (*rank_grpc.RankEmptyResponse, error) {
	if err := server.svc.RankTopKUser(ctx); err != nil {
		return &rank_grpc.RankEmptyResponse{}, err
	}
	return &rank_grpc.RankEmptyResponse{}, nil
}

func (server *RankServiceServer) RankTopKPost(ctx context.Context, req *rank_grpc.RankEmptyRequest) (*rank_grpc.RankEmptyResponse, error) {
	if err := server.svc.RankTopKPost(ctx); err != nil {
		return &rank_grpc.RankEmptyResponse{}, err
	}
	return &rank_grpc.RankEmptyResponse{}, nil
}

func (server *RankServiceServer) TopKUser(ctx context.Context, req *rank_grpc.RankEmptyRequest) (*rank_grpc.TopKUserResponse, error) {
	users, err := server.svc.TopKUser(ctx)
	if err != nil {
		return &rank_grpc.TopKUserResponse{}, err
	}

	respUsers := make([]*rank_grpc.RankUser, 0, len(users))
	for _, user := range users {
		respUsers = append(respUsers, &rank_grpc.RankUser{
			ID:    user.ID,
			Score: user.Score,
		})
	}

	return &rank_grpc.TopKUserResponse{Users: respUsers}, nil
}

func (server *RankServiceServer) TopKPost(ctx context.Context, req *rank_grpc.RankEmptyRequest) (*rank_grpc.TopKPostResponse, error) {
	posts, err := server.svc.TopKPost(ctx)
	if err != nil {
		return &rank_grpc.TopKPostResponse{}, err
	}

	respPosts := make([]*rank_grpc.RankPost, 0, len(posts))
	for _, post := range posts {
		respPosts = append(respPosts, &rank_grpc.RankPost{
			ID:    post.ID,
			Score: post.Score,
		})
	}

	return &rank_grpc.TopKPostResponse{Posts: respPosts}, nil
}

func (server *RankServiceServer) HealthCheck(ctx context.Context, req *rank_grpc.HealthCheckRequest) (*rank_grpc.HealthCheckResponse, error) {
	return &rank_grpc.HealthCheckResponse{}, nil
}
