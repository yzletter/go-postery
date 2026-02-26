package grpc

import (
	"context"

	auth_grpc "github.com/yzletter/go-postery/api/proto/auth/v1"
	"github.com/yzletter/go-postery/auth/service"
	"github.com/yzletter/go-postery/auth/utils"
)

type AuthServiceServer struct {
	svc service.AuthService
	auth_grpc.UnimplementedAuthServiceServer
}

func NewAuthServiceServer(svc service.AuthService) *AuthServiceServer {
	return &AuthServiceServer{
		svc: svc,
	}
}

func (server *AuthServiceServer) LoginByPassword(ctx context.Context, req *auth_grpc.LoginByPasswordRequest) (*auth_grpc.UserID, error) {
	// 调用 Service
	uid, err := server.svc.LoginByPassword(ctx, req.Identifier, req.Password)
	if err != nil {
		return &auth_grpc.UserID{}, err
	}

	// 返回 Response
	return &auth_grpc.UserID{UserID: uid}, nil
}

func (server *AuthServiceServer) LoginByPhone(ctx context.Context, req *auth_grpc.LoginByPhoneRequest) (*auth_grpc.UserID, error) {
	// 调用 Service
	uid, err := server.svc.LoginByPhone(ctx, req.Phone, req.Code)
	if err != nil {
		return &auth_grpc.UserID{}, err
	}
	// 返回 Response
	return &auth_grpc.UserID{UserID: uid}, nil
}

func (server *AuthServiceServer) HasPassword(ctx context.Context, id *auth_grpc.UserID) (*auth_grpc.HasPasswordResponse, error) {
	// 调用 Service
	has, err := server.svc.HasPassword(ctx, id.UserID)
	if err != nil {
		return &auth_grpc.HasPasswordResponse{}, err
	}
	// 返回 Response
	return &auth_grpc.HasPasswordResponse{Result: has}, nil
}

func (server *AuthServiceServer) SetPassword(ctx context.Context, req *auth_grpc.SetPasswordRequest) (*auth_grpc.AuthEmptyResponse, error) {
	// 调用 Service
	if err := server.svc.SetPassword(ctx, req.UserID, req.Code, req.Password); err != nil {
		return &auth_grpc.AuthEmptyResponse{}, err
	}
	// 返回 Response
	return &auth_grpc.AuthEmptyResponse{}, nil
}

func (server *AuthServiceServer) UpdatePassword(ctx context.Context, req *auth_grpc.UpdatePasswordRequest) (*auth_grpc.AuthEmptyResponse, error) {
	// 调用 Service
	if err := server.svc.UpdatePassword(ctx, req.UserID, req.OldPassword, req.NewPassword); err != nil {
		return &auth_grpc.AuthEmptyResponse{}, err
	}
	// 返回 Response
	return &auth_grpc.AuthEmptyResponse{}, nil
}

func (server *AuthServiceServer) GetAuthIdentityByUID(ctx context.Context, id *auth_grpc.UserID) (*auth_grpc.AuthIdentity, error) {
	// 调用 Service
	phone, email, err := server.svc.GetAuthIdentityByUID(ctx, id.UserID)
	if err != nil {
		return &auth_grpc.AuthIdentity{}, err
	}
	// 返回 Response
	return &auth_grpc.AuthIdentity{Phone: phone, Email: email}, nil
}

func (server *AuthServiceServer) IssueTokens(ctx context.Context, req *auth_grpc.IssueTokenRequest) (*auth_grpc.DualTokens, error) {
	// 调用 Service
	accessToken, refreshToken, err := server.svc.IssueTokens(ctx, req.UserID, int(req.Role), req.UserAgent)
	if err != nil {
		return &auth_grpc.DualTokens{}, err
	}
	// 返回 Response
	return &auth_grpc.DualTokens{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (server *AuthServiceServer) ClearTokens(ctx context.Context, tokens *auth_grpc.DualTokens) (*auth_grpc.AuthEmptyResponse, error) {
	// 调用 Service
	if err := server.svc.ClearTokens(ctx, tokens.AccessToken, tokens.RefreshToken); err != nil {
		return &auth_grpc.AuthEmptyResponse{}, err
	}
	// 返回 Response
	return &auth_grpc.AuthEmptyResponse{}, nil
}

func (server *AuthServiceServer) VerifyAccessToken(ctx context.Context, token *auth_grpc.AccessToken) (*auth_grpc.JWTTokenClaims, error) {
	// 调用 Service
	claim, err := server.svc.VerifyAccessToken(ctx, token.AccessToken)
	if err != nil {
		return &auth_grpc.JWTTokenClaims{}, err
	}
	// 返回 Response
	return &auth_grpc.JWTTokenClaims{
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

func (server *AuthServiceServer) GetInfoByRefreshToken(ctx context.Context, token *auth_grpc.RefreshToken) (*auth_grpc.GetInfoByRefreshTokenResponse, error) {
	// 调用 Service
	uid, role, ssid, err := server.svc.GetInfoByRefreshToken(ctx, token.RefreshToken)
	if err != nil {
		return &auth_grpc.GetInfoByRefreshTokenResponse{}, err
	}
	// 返回 Response
	return &auth_grpc.GetInfoByRefreshTokenResponse{UserID: uid, Role: int32(role), SSID: ssid}, nil
}

func (server *AuthServiceServer) CheckBlackList(ctx context.Context, req *auth_grpc.CheckBlackListRequest) (*auth_grpc.CheckBlackListResponse, error) {
	// 调用 Service
	exist, err := server.svc.CheckBlackList(ctx, req.SSID)
	if err != nil {
		return &auth_grpc.CheckBlackListResponse{}, err
	}
	// 返回 Response
	return &auth_grpc.CheckBlackListResponse{Result: exist}, nil
}
