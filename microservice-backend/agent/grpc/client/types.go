package client

import (
	"context"

	post_grpc "github.com/yzletter/go-postery/api/proto/post/v1"
	hub2 "github.com/yzletter/go-postery/microservice-backend/agent/grpc/hub"
)

type ServiceHub interface {
	LoadEndpoints(ctx context.Context, service string)
	WatchEndpointsFromServiceHub(ctx context.Context, service string)
	Take(ctx context.Context, service string) *hub2.Endpoint
}

const (
	PostServiceName = "post_service"
)

type PostClient interface {
	GetDetailByID(ctx context.Context, req *post_grpc.GetDetailByIDRequest) (*post_grpc.PostDetail, error)
	Close()
}
