package client

import (
	"context"

	auth_grpc "github.com/yzletter/go-postery/api/proto/auth/v1"
	code_grpc "github.com/yzletter/go-postery/api/proto/code/v1"
)

const (
	CodeClientAddr    = "172.16.150.246:9001"
	AuthClientAddr    = "172.16.150.246:9002"
	LotteryClientAddr = "172.16.150.246:9003"
	PostClientAddr    = "172.16.150.246:9004"
)

type CodeClient interface {
	Send(ctx context.Context, req *code_grpc.SendCodeRequest) (*code_grpc.SendCodeResponse, error)
	Verify(ctx context.Context, req *code_grpc.CheckCodeRequest) (*code_grpc.CheckCodeResponse, error)
	Close()
}

type AuthClient interface {
	LoginByPassword(ctx context.Context, req *auth_grpc.LoginByPasswordRequest) (*auth_grpc.UserID, error)
	LoginByPhone(ctx context.Context, req *auth_grpc.LoginByPhoneRequest) (*auth_grpc.UserID, error)
	HasPassword(ctx context.Context, req *auth_grpc.UserID) (*auth_grpc.HasPasswordResponse, error)
	SetPassword(ctx context.Context, req *auth_grpc.SetPasswordRequest) (*auth_grpc.AuthEmptyResponse, error)
	UpdatePassword(ctx context.Context, req *auth_grpc.UpdatePasswordRequest) (*auth_grpc.AuthEmptyResponse, error)
	GetAuthIdentityByUID(ctx context.Context, req *auth_grpc.UserID) (*auth_grpc.AuthIdentity, error)
	IssueTokens(ctx context.Context, req *auth_grpc.IssueTokenRequest) (*auth_grpc.DualTokens, error)
	ClearTokens(ctx context.Context, req *auth_grpc.DualTokens) (*auth_grpc.AuthEmptyResponse, error)
	VerifyAccessToken(ctx context.Context, req *auth_grpc.AccessToken) (*auth_grpc.JWTTokenClaims, error)
	GetInfoByRefreshToken(ctx context.Context, req *auth_grpc.RefreshToken) (*auth_grpc.GetInfoByRefreshTokenResponse, error)
	CheckBlackList(ctx context.Context, req *auth_grpc.CheckBlackListRequest) (*auth_grpc.CheckBlackListResponse, error)
	Close()
}
