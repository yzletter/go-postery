package client

import (
	"context"

	post_grpc "github.com/yzletter/go-postery/api/proto/post/v1"
)

type ServiceHub interface {
	GetServiceEndpoint(ctx context.Context, service string) string
}

const (
	PostServiceName = "post_service"
)

type PostClient interface {
	GetDetailByID(ctx context.Context, req *post_grpc.GetDetailByIDRequest) (*post_grpc.PostDetail, error)
	Close()
}
