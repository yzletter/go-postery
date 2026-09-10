package grpc

import (
	"context"

	rank_grpc2 "github.com/yzletter/go-postery/backend/api/proto/rank/v1"
	"github.com/yzletter/go-postery/backend/micro/rank/service"
)

type RankServiceServer struct {
	svc service.RankService
	rank_grpc2.UnimplementedRankServiceServer
}

func NewRankServiceServer(svc service.RankService) *RankServiceServer {
	return &RankServiceServer{
		svc: svc,
	}
}

func (server *RankServiceServer) RankUser(ctx context.Context, req *rank_grpc2.RankIDRequest) (*rank_grpc2.RankEmptyResponse, error) {
	if err := server.svc.RankUser(ctx, req.ID); err != nil {
		return &rank_grpc2.RankEmptyResponse{}, err
	}
	return &rank_grpc2.RankEmptyResponse{}, nil
}

func (server *RankServiceServer) RankPost(ctx context.Context, req *rank_grpc2.RankIDRequest) (*rank_grpc2.RankEmptyResponse, error) {
	if err := server.svc.RankPost(ctx, req.ID); err != nil {
		return &rank_grpc2.RankEmptyResponse{}, err
	}
	return &rank_grpc2.RankEmptyResponse{}, nil
}

func (server *RankServiceServer) RankTopKUser(ctx context.Context, req *rank_grpc2.RankEmptyRequest) (*rank_grpc2.RankEmptyResponse, error) {
	if err := server.svc.RankTopKUser(ctx); err != nil {
		return &rank_grpc2.RankEmptyResponse{}, err
	}
	return &rank_grpc2.RankEmptyResponse{}, nil
}

func (server *RankServiceServer) RankTopKPost(ctx context.Context, req *rank_grpc2.RankEmptyRequest) (*rank_grpc2.RankEmptyResponse, error) {
	if err := server.svc.RankTopKPost(ctx); err != nil {
		return &rank_grpc2.RankEmptyResponse{}, err
	}
	return &rank_grpc2.RankEmptyResponse{}, nil
}

func (server *RankServiceServer) TopKUser(ctx context.Context, req *rank_grpc2.RankEmptyRequest) (*rank_grpc2.TopKUserResponse, error) {
	users, err := server.svc.TopKUser(ctx)
	if err != nil {
		return &rank_grpc2.TopKUserResponse{}, err
	}

	respUsers := make([]*rank_grpc2.RankUser, 0, len(users))
	for _, user := range users {
		respUsers = append(respUsers, &rank_grpc2.RankUser{
			ID:    user.ID,
			Score: user.Score,
		})
	}

	return &rank_grpc2.TopKUserResponse{Users: respUsers}, nil
}

func (server *RankServiceServer) TopKPost(ctx context.Context, req *rank_grpc2.RankEmptyRequest) (*rank_grpc2.TopKPostResponse, error) {
	posts, err := server.svc.TopKPost(ctx)
	if err != nil {
		return &rank_grpc2.TopKPostResponse{}, err
	}

	respPosts := make([]*rank_grpc2.RankPost, 0, len(posts))
	for _, post := range posts {
		respPosts = append(respPosts, &rank_grpc2.RankPost{
			ID:    post.ID,
			Score: post.Score,
		})
	}

	return &rank_grpc2.TopKPostResponse{Posts: respPosts}, nil
}

func (server *RankServiceServer) HealthCheck(ctx context.Context, req *rank_grpc2.HealthCheckRequest) (*rank_grpc2.HealthCheckResponse, error) {
	return &rank_grpc2.HealthCheckResponse{}, nil
}
