package grpc

import (
	"context"

	auth_grpc2 "github.com/yzletter/go-postery/backend/api/proto/auth/v1"
	"github.com/yzletter/go-postery/backend/micro/auth/service"
	"github.com/yzletter/go-postery/backend/utils"
)

type AuthServiceServer struct {
	svc service.AuthService
	auth_grpc2.UnimplementedAuthServiceServer
}

func NewAuthServiceServer(svc service.AuthService) *AuthServiceServer {
	return &AuthServiceServer{
		svc: svc,
	}
}

func (server *AuthServiceServer) Login(ctx context.Context, req *auth_grpc2.LoginRequest) (*auth_grpc2.LoginResponse, error) {
	var uid int64
	var err error

	// 调用 Service
	switch req.Biz {
	case auth_grpc2.LoginBiz_LOGIN_BIZ_Phone:
		uid, err = server.svc.LoginByPhone(ctx, req.Identifier, req.Verification)
	case auth_grpc2.LoginBiz_LOGIN_BIZ_Password:
		uid, err = server.svc.LoginByPassword(ctx, req.Identifier, req.Verification)
	}
	if err != nil {
		return &auth_grpc2.LoginResponse{}, err
	}

	// 返回 Response
	return &auth_grpc2.LoginResponse{UserID: uid}, nil
}

func (server *AuthServiceServer) HasPassword(ctx context.Context, id *auth_grpc2.UserID) (*auth_grpc2.HasPasswordResponse, error) {
	// 调用 Service
	has, err := server.svc.HasPassword(ctx, id.UserID)
	if err != nil {
		return &auth_grpc2.HasPasswordResponse{}, err
	}
	// 返回 Response
	return &auth_grpc2.HasPasswordResponse{Result: has}, nil
}

func (server *AuthServiceServer) SetPassword(ctx context.Context, req *auth_grpc2.SetPasswordRequest) (*auth_grpc2.AuthEmptyResponse, error) {
	// 调用 Service
	if err := server.svc.SetPassword(ctx, req.UserID, req.Code, req.Password); err != nil {
		return &auth_grpc2.AuthEmptyResponse{}, err
	}
	// 返回 Response
	return &auth_grpc2.AuthEmptyResponse{}, nil
}

func (server *AuthServiceServer) UpdatePassword(ctx context.Context, req *auth_grpc2.UpdatePasswordRequest) (*auth_grpc2.AuthEmptyResponse, error) {
	// 调用 Service
	if err := server.svc.UpdatePassword(ctx, req.UserID, req.OldPassword, req.NewPassword); err != nil {
		return &auth_grpc2.AuthEmptyResponse{}, err
	}
	// 返回 Response
	return &auth_grpc2.AuthEmptyResponse{}, nil
}

func (server *AuthServiceServer) GetAuthIdentityByUID(ctx context.Context, id *auth_grpc2.UserID) (*auth_grpc2.AuthIdentity, error) {
	// 调用 Service
	phone, email, err := server.svc.GetAuthIdentityByUID(ctx, id.UserID)
	if err != nil {
		return &auth_grpc2.AuthIdentity{}, err
	}
	// 返回 Response
	return &auth_grpc2.AuthIdentity{Phone: phone, Email: email}, nil
}

func (server *AuthServiceServer) IssueTokens(ctx context.Context, req *auth_grpc2.IssueTokenRequest) (*auth_grpc2.DualTokens, error) {
	// 调用 Service
	accessToken, refreshToken, err := server.svc.IssueTokens(ctx, req.UserID, int(req.Role), req.UserAgent)
	if err != nil {
		return &auth_grpc2.DualTokens{}, err
	}
	// 返回 Response
	return &auth_grpc2.DualTokens{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (server *AuthServiceServer) ClearTokens(ctx context.Context, tokens *auth_grpc2.DualTokens) (*auth_grpc2.AuthEmptyResponse, error) {
	// 调用 Service
	if err := server.svc.ClearTokens(ctx, tokens.AccessToken, tokens.RefreshToken); err != nil {
		return &auth_grpc2.AuthEmptyResponse{}, err
	}
	// 返回 Response
	return &auth_grpc2.AuthEmptyResponse{}, nil
}

func (server *AuthServiceServer) VerifyAccessToken(ctx context.Context, token *auth_grpc2.AccessToken) (*auth_grpc2.JWTTokenClaims, error) {
	// 调用 Service
	claim, err := server.svc.VerifyAccessToken(ctx, token.AccessToken)
	if err != nil {
		return &auth_grpc2.JWTTokenClaims{}, err
	}
	// 返回 Response
	return &auth_grpc2.JWTTokenClaims{
		UserID:    claim.Uid,
		SSID:      claim.SSid,
		Role:      int32(claim.Role),
		UserAgent: claim.UserAgent,
		Issuer:    claim.Issuer,
		Subject:   claim.Subject,
		Audience:  claim.Audience,
		ExpiresAt: utils.GoTimeToRPCTime(claim.ExpiresAt),
		NotBefore: utils.GoTimeToRPCTime(claim.NotBefore),
		IssuedAt:  utils.GoTimeToRPCTime(claim.IssuedAt),
		ID:        claim.ID,
	}, nil
}

func (server *AuthServiceServer) GetInfoByRefreshToken(ctx context.Context, token *auth_grpc2.RefreshToken) (*auth_grpc2.GetInfoByRefreshTokenResponse, error) {
	// 调用 Service
	uid, role, ssid, err := server.svc.GetInfoByRefreshToken(ctx, token.RefreshToken)
	if err != nil {
		return &auth_grpc2.GetInfoByRefreshTokenResponse{}, err
	}
	// 返回 Response
	return &auth_grpc2.GetInfoByRefreshTokenResponse{UserID: uid, Role: int32(role), SSID: ssid}, nil
}

func (server *AuthServiceServer) CheckBlackList(ctx context.Context, req *auth_grpc2.CheckBlackListRequest) (*auth_grpc2.CheckBlackListResponse, error) {
	// 调用 Service
	exist, err := server.svc.CheckBlackList(ctx, req.SSID)
	if err != nil {
		return &auth_grpc2.CheckBlackListResponse{}, err
	}
	// 返回 Response
	return &auth_grpc2.CheckBlackListResponse{Result: exist}, nil
}

func (server *AuthServiceServer) HealthCheck(ctx context.Context, req *auth_grpc2.HealthCheckRequest) (*auth_grpc2.HealthCheckResponse, error) {
	return &auth_grpc2.HealthCheckResponse{}, nil
}
