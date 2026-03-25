package client

import (
	"context"
	"time"

	auth_grpc "github.com/yzletter/go-postery/api/proto/auth/v1"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

type authClient struct {
	conn   *grpc.ClientConn
	client auth_grpc.AuthServiceClient
}

func NewAuthClient() (AuthClient, error) {
	// 建议：启用 ka，避免中间网络设备把长连接静默掐掉
	ka := keepalive.ClientParameters{
		Time:                30 * time.Second,
		Timeout:             10 * time.Second,
		PermitWithoutStream: true,
	}

	conn, err := grpc.NewClient(
		AuthClientAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()), // 生产用 TLS
		CircuitBreakerDialOption(),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()), // Jaeger
		grpc.WithKeepaliveParams(ka),
	)
	if err != nil {
		return nil, err
	}

	return &authClient{
		conn:   conn,
		client: auth_grpc.NewAuthServiceClient(conn),
	}, nil
}

func (client *authClient) Close() {
	_ = client.conn.Close()
}

func (client *authClient) LoginByPassword(ctx context.Context, req *auth_grpc.LoginByPasswordRequest) (*auth_grpc.UserID, error) {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(ctx, 8000*time.Millisecond)
	defer cancel()

	return client.client.LoginByPassword(ctx, req)
}

func (client *authClient) LoginByPhone(ctx context.Context, req *auth_grpc.LoginByPhoneRequest) (*auth_grpc.UserID, error) {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(ctx, 8000*time.Millisecond)
	defer cancel()

	return client.client.LoginByPhone(ctx, req)
}

func (client *authClient) HasPassword(ctx context.Context, req *auth_grpc.UserID) (*auth_grpc.HasPasswordResponse, error) {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(ctx, 8000*time.Millisecond)
	defer cancel()

	return client.client.HasPassword(ctx, req)
}

func (client *authClient) SetPassword(ctx context.Context, req *auth_grpc.SetPasswordRequest) (*auth_grpc.AuthEmptyResponse, error) {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(ctx, 8000*time.Millisecond)
	defer cancel()

	return client.client.SetPassword(ctx, req)
}

func (client *authClient) UpdatePassword(ctx context.Context, req *auth_grpc.UpdatePasswordRequest) (*auth_grpc.AuthEmptyResponse, error) {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(ctx, 8000*time.Millisecond)
	defer cancel()

	return client.client.UpdatePassword(ctx, req)
}

func (client *authClient) GetAuthIdentityByUID(ctx context.Context, req *auth_grpc.UserID) (*auth_grpc.AuthIdentity, error) {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(ctx, 8000*time.Millisecond)
	defer cancel()

	return client.client.GetAuthIdentityByUID(ctx, req)
}

func (client *authClient) IssueTokens(ctx context.Context, req *auth_grpc.IssueTokenRequest) (*auth_grpc.DualTokens, error) {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(ctx, 8000*time.Millisecond)
	defer cancel()

	return client.client.IssueTokens(ctx, req)
}

func (client *authClient) ClearTokens(ctx context.Context, req *auth_grpc.DualTokens) (*auth_grpc.AuthEmptyResponse, error) {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(ctx, 8000*time.Millisecond)
	defer cancel()

	return client.client.ClearTokens(ctx, req)
}

func (client *authClient) VerifyAccessToken(ctx context.Context, req *auth_grpc.AccessToken) (*auth_grpc.JWTTokenClaims, error) {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(ctx, 8000*time.Millisecond)
	defer cancel()

	return client.client.VerifyAccessToken(ctx, req)
}

func (client *authClient) GetInfoByRefreshToken(ctx context.Context, req *auth_grpc.RefreshToken) (*auth_grpc.GetInfoByRefreshTokenResponse, error) {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(ctx, 8000*time.Millisecond)
	defer cancel()

	return client.client.GetInfoByRefreshToken(ctx, req)
}

func (client *authClient) CheckBlackList(ctx context.Context, req *auth_grpc.CheckBlackListRequest) (*auth_grpc.CheckBlackListResponse, error) {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(ctx, 8000*time.Millisecond)
	defer cancel()

	return client.client.CheckBlackList(ctx, req)
}
