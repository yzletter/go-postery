package client

import (
	"context"
	"time"

	user_grpc "github.com/yzletter/go-postery/api/proto/user/v1"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

type userClient struct {
	conn   *grpc.ClientConn
	client user_grpc.UserServiceClient
}

func NewUserClient() (UserClient, error) {
	// 建议：启用 ka，避免中间网络设备把长连接静默掐掉
	ka := keepalive.ClientParameters{
		Time:                30 * time.Second,
		Timeout:             10 * time.Second,
		PermitWithoutStream: true,
	}

	conn, err := grpc.NewClient(
		UserClientAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()), // 生产用 TLS
		CircuitBreakerDialOption(),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()), // Jaeger
		grpc.WithKeepaliveParams(ka),
	)
	if err != nil {
		return nil, err
	}

	return &userClient{
		conn:   conn,
		client: user_grpc.NewUserServiceClient(conn),
	}, nil
}

func (client *userClient) Close() {
	_ = client.conn.Close()
}

func (client *userClient) GetProfileById(ctx context.Context, req *user_grpc.GetProfileByIdRequest) (*user_grpc.UserDetail, error) {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(ctx, 8000*time.Millisecond)
	defer cancel()

	return client.client.GetProfileById(ctx, req)
}

func (client *userClient) UpdateProfile(ctx context.Context, req *user_grpc.UpdateProfileRequest) (*user_grpc.UpdateProfileResponse, error) {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(ctx, 8000*time.Millisecond)
	defer cancel()

	return client.client.UpdateProfile(ctx, req)
}

func (client *userClient) Top(ctx context.Context, req *user_grpc.TopRequest) (*user_grpc.TopResponse, error) {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(ctx, 8000*time.Millisecond)
	defer cancel()

	return client.client.Top(ctx, req)
}

func (client *userClient) Follow(ctx context.Context, req *user_grpc.FollowCommonRequest) (*user_grpc.FollowEmptyResponse, error) {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(ctx, 8000*time.Millisecond)
	defer cancel()

	return client.client.Follow(ctx, req)
}

func (client *userClient) UnFollow(ctx context.Context, req *user_grpc.FollowCommonRequest) (*user_grpc.FollowEmptyResponse, error) {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(ctx, 8000*time.Millisecond)
	defer cancel()

	return client.client.UnFollow(ctx, req)
}

func (client *userClient) IfFollow(ctx context.Context, req *user_grpc.FollowCommonRequest) (*user_grpc.IfFollowResponse, error) {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(ctx, 8000*time.Millisecond)
	defer cancel()

	return client.client.IfFollow(ctx, req)
}

func (client *userClient) ListFollowersByPage(ctx context.Context, req *user_grpc.ListFollowRequest) (*user_grpc.ListFollowResponse, error) {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(ctx, 8000*time.Millisecond)
	defer cancel()

	return client.client.ListFollowersByPage(ctx, req)
}

func (client *userClient) ListFolloweesByPage(ctx context.Context, req *user_grpc.ListFollowRequest) (*user_grpc.ListFollowResponse, error) {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(ctx, 8000*time.Millisecond)
	defer cancel()

	return client.client.ListFolloweesByPage(ctx, req)
}
