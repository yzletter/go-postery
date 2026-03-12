package client

import (
	"context"
	"time"

	session_grpc "github.com/yzletter/go-postery/api/proto/session/v1"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

type sessionClient struct {
	conn   *grpc.ClientConn
	client session_grpc.SessionServiceClient
}

func NewSessionClient() (SessionClient, error) {
	// 建议：启用 ka，避免中间网络设备把长连接静默掐掉
	ka := keepalive.ClientParameters{
		Time:                30 * time.Second,
		Timeout:             10 * time.Second,
		PermitWithoutStream: true,
	}

	conn, err := grpc.NewClient(
		SessionClientAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()), // 生产用 TLS
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),       // Jaeger
		grpc.WithKeepaliveParams(ka),
	)
	if err != nil {
		return nil, err
	}

	return &sessionClient{
		conn:   conn,
		client: session_grpc.NewSessionServiceClient(conn),
	}, nil
}

func (client *sessionClient) Close() {
	_ = client.conn.Close()
}

func (client *sessionClient) ListByUID(ctx context.Context, req *session_grpc.UserID) (*session_grpc.Sessions, error) {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(ctx, 8000*time.Millisecond)
	defer cancel()

	return client.client.ListByUID(ctx, req)
}

func (client *sessionClient) GetSession(ctx context.Context, req *session_grpc.BothUserID) (*session_grpc.Session, error) {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(ctx, 8000*time.Millisecond)
	defer cancel()

	return client.client.GetSession(ctx, req)
}

func (client *sessionClient) GetHistoryMessagesByPage(ctx context.Context, req *session_grpc.GetHistoryMessagesByPageRequest) (*session_grpc.GetHistoryMessagesByPageResponse, error) {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(ctx, 8000*time.Millisecond)
	defer cancel()

	return client.client.GetHistoryMessagesByPage(ctx, req)
}

func (client *sessionClient) Delete(ctx context.Context, req *session_grpc.DeleteRequest) (*session_grpc.SessionEmptyResponse, error) {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(ctx, 8000*time.Millisecond)
	defer cancel()

	return client.client.Delete(ctx, req)
}

func (client *sessionClient) UpdateUnread(ctx context.Context, req *session_grpc.UpdateUnreadRequest) (*session_grpc.SessionEmptyResponse, error) {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(ctx, 8000*time.Millisecond)
	defer cancel()

	return client.client.UpdateUnread(ctx, req)
}

func (client *sessionClient) ClearUnread(ctx context.Context, req *session_grpc.ClearUnreadRequest) (*session_grpc.SessionEmptyResponse, error) {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(ctx, 8000*time.Millisecond)
	defer cancel()

	return client.client.ClearUnread(ctx, req)
}

func (client *sessionClient) CreateMessage(ctx context.Context, req *session_grpc.Message) (*session_grpc.Message, error) {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(ctx, 8000*time.Millisecond)
	defer cancel()

	return client.client.CreateMessage(ctx, req)
}
