package client

import (
	"context"

	post_grpc "github.com/yzletter/go-postery/api/proto/post/v1"
	"google.golang.org/grpc"
)

type postClient struct {
	conn   *grpc.ClientConn
	client post_grpc.PostServiceClient
}

func NewPostClient(conn *grpc.ClientConn) (PostClient, error) {
	if err := validateConn(conn); err != nil {
		return nil, err
	}

	return &postClient{
		conn:   conn,
		client: post_grpc.NewPostServiceClient(conn),
	}, nil
}

func (client *postClient) Close() {
	_ = client.conn.Close()
}

func (client *postClient) GetDetailByID(ctx context.Context, req *post_grpc.GetDetailByIDRequest) (*post_grpc.PostDetail, error) {
	return client.client.GetDetailByID(ctx, req)
}
