package service

import (
	"context"

	auth_grpc "github.com/yzletter/go-postery/api/proto/auth/v1"
)

type AuthService interface {
	LoginByPassword(context.Context, *auth_grpc.LoginByPasswordRequest) (*auth_grpc.UserID, error)
	LoginByPhone(context.Context, *auth_grpc.LoginByPhoneRequest) (*auth_grpc.UserID, error)
	HasPassword(context.Context, *auth_grpc.UserID) (*auth_grpc.HasPasswordResponse, error)
	SetPassword(context.Context, *auth_grpc.SetPasswordRequest) (*auth_grpc.AuthEmptyResponse, error)
	UpdatePassword(context.Context, *auth_grpc.UpdatePasswordRequest) (*auth_grpc.AuthEmptyResponse, error)
	GetAuthIdentityByUID(context.Context, *auth_grpc.UserID) (*auth_grpc.AuthIdentity, error)
	IssueTokens(context.Context, *auth_grpc.IssueTokenRequest) (*auth_grpc.DualTokens, error)
	ClearTokens(context.Context, *auth_grpc.DualTokens) (*auth_grpc.AuthEmptyResponse, error)
	VerifyAccessToken(context.Context, *auth_grpc.AccessToken) (*auth_grpc.JWTTokenClaims, error)
	GetInfoByRefreshToken(context.Context, *auth_grpc.RefreshToken) (*auth_grpc.GetInfoByRefreshTokenResponse, error)
	CheckBlackList(context.Context, *auth_grpc.CheckBlackListRequest) (*auth_grpc.CheckBlackListResponse, error)
	auth_grpc.UnsafeAuthServiceServer
}
