package client

import (
	"context"
	"time"

	code_grpc "github.com/yzletter/go-postery/api/proto/code/v1"
	"google.golang.org/grpc"
)

type codeClient struct {
	conn   *grpc.ClientConn
	client code_grpc.CodeServiceClient
}

func NewCodeClient(conn *grpc.ClientConn) (CodeClient, error) {
	if err := validateConn(conn); err != nil {
		return nil, err
	}

	return &codeClient{
		conn:   conn,
		client: code_grpc.NewCodeServiceClient(conn),
	}, nil
}

func (client *codeClient) Close() {
	_ = client.conn.Close()
}

func (client *codeClient) Send(ctx context.Context, req *code_grpc.SendCodeRequest) (*code_grpc.SendCodeResponse, error) {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(ctx, 8000*time.Millisecond)
	defer cancel()

	return client.client.Send(ctx, req)
}

func (client *codeClient) Verify(ctx context.Context, req *code_grpc.CheckCodeRequest) (*code_grpc.CheckCodeResponse, error) {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(ctx, 8000*time.Millisecond)
	defer cancel()

	return client.client.Verify(ctx, req)
}
