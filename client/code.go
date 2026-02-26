package client

import (
	"context"
	"time"

	code_grpc "github.com/yzletter/go-postery/api/proto/code/v1"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

type codeClient struct {
	conn   *grpc.ClientConn
	client code_grpc.CodeServiceClient
}

func NewCodeClient(target string) (CodeClient, error) {
	// 建议：启用 ka，避免中间网络设备把长连接静默掐掉
	ka := keepalive.ClientParameters{
		Time:                30 * time.Second,
		Timeout:             10 * time.Second,
		PermitWithoutStream: true,
	}

	conn, err := grpc.NewClient(
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()), // 生产用 TLS
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),       // Jaeger
		grpc.WithKeepaliveParams(ka),
	)
	if err != nil {
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
	ctx, cancel := context.WithTimeout(ctx, 800*time.Millisecond)
	defer cancel()

	return client.client.Send(ctx, req)
}

func (client *codeClient) Verify(ctx context.Context, req *code_grpc.CheckCodeRequest) (*code_grpc.CheckCodeResponse, error) {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(ctx, 800*time.Millisecond)
	defer cancel()

	return client.client.Verify(ctx, req)
}
