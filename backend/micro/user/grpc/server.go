package grpc

import (
	"context"

	user_grpc "github.com/yzletter/go-postery/api/proto/user/v1"
	dto2 "github.com/yzletter/go-postery/backend/micro/user/dto"
	"github.com/yzletter/go-postery/backend/micro/user/service"
)

type UserServiceServer struct {
	svc service.UserService
	user_grpc.UnimplementedUserServiceServer
}

func NewUserServiceServer(svc service.UserService) *UserServiceServer {
	return &UserServiceServer{
		svc: svc,
	}
}

func (server *UserServiceServer) GetProfileById(ctx context.Context, req *user_grpc.GetProfileByIdRequest) (*user_grpc.UserDetail, error) {
	profile, err := server.svc.GetProfileByID(ctx, req.ID)
	if err != nil {
		return &user_grpc.UserDetail{}, err
	}
	return dto2.ToUserDetail(profile), nil
}

func (server *UserServiceServer) UpdateProfile(ctx context.Context, req *user_grpc.UpdateProfileRequest) (*user_grpc.UpdateProfileResponse, error) {
	profile := dto2.UpdateProfileRequestToModel(req)
	profile.UserID = req.ID

	if err := server.svc.UpdateProfile(ctx, profile); err != nil {
		return &user_grpc.UpdateProfileResponse{}, err
	}
	return &user_grpc.UpdateProfileResponse{}, nil
}

func (server *UserServiceServer) Top(ctx context.Context, req *user_grpc.TopRequest) (*user_grpc.TopResponse, error) {
	profiles, scores, err := server.svc.Top(ctx)
	if err != nil {
		return &user_grpc.TopResponse{}, err
	}

	topUsers := make([]*user_grpc.TopUser, 0, len(profiles))
	for idx, profile := range profiles {
		topUsers = append(topUsers, dto2.ToTopUser(profile, scores[idx]))
	}

	return &user_grpc.TopResponse{TopUsers: topUsers}, nil
}

func (server *UserServiceServer) Follow(ctx context.Context, req *user_grpc.FollowCommonRequest) (*user_grpc.FollowEmptyResponse, error) {
	if err := server.svc.Follow(ctx, req.FollowerID, req.FolloweeID); err != nil {
		return &user_grpc.FollowEmptyResponse{}, err
	}
	return &user_grpc.FollowEmptyResponse{}, nil
}

func (server *UserServiceServer) UnFollow(ctx context.Context, req *user_grpc.FollowCommonRequest) (*user_grpc.FollowEmptyResponse, error) {
	if err := server.svc.UnFollow(ctx, req.FollowerID, req.FolloweeID); err != nil {
		return &user_grpc.FollowEmptyResponse{}, err
	}
	return &user_grpc.FollowEmptyResponse{}, nil
}

func (server *UserServiceServer) IfFollow(ctx context.Context, req *user_grpc.FollowCommonRequest) (*user_grpc.IfFollowResponse, error) {
	result, err := server.svc.IfFollow(ctx, req.FollowerID, req.FolloweeID)
	if err != nil {
		return &user_grpc.IfFollowResponse{Result: -1}, err
	}
	return &user_grpc.IfFollowResponse{Result: int32(result)}, nil
}

func (server *UserServiceServer) ListFollowersByPage(ctx context.Context, req *user_grpc.ListFollowRequest) (*user_grpc.ListFollowResponse, error) {
	total, profiles, err := server.svc.ListFollowersByPage(ctx, req.UserID, int(req.PageNo), int(req.PageSize))
	if err != nil {
		return &user_grpc.ListFollowResponse{}, err
	}

	userBriefs := make([]*user_grpc.UserBrief, 0, len(profiles))
	for _, profile := range profiles {
		userBriefs = append(userBriefs, dto2.ToUserBrief(profile))
	}

	return &user_grpc.ListFollowResponse{Count: uint64(total), UserBriefs: userBriefs}, nil
}

func (server *UserServiceServer) ListFolloweesByPage(ctx context.Context, req *user_grpc.ListFollowRequest) (*user_grpc.ListFollowResponse, error) {
	total, profiles, err := server.svc.ListFolloweesByPage(ctx, req.UserID, int(req.PageNo), int(req.PageSize))
	if err != nil {
		return &user_grpc.ListFollowResponse{}, err
	}

	userBriefs := make([]*user_grpc.UserBrief, 0, len(profiles))
	for _, profile := range profiles {
		userBriefs = append(userBriefs, dto2.ToUserBrief(profile))
	}

	return &user_grpc.ListFollowResponse{Count: uint64(total), UserBriefs: userBriefs}, nil
}

func (server *UserServiceServer) UploadAvatarSign(ctx context.Context, req *user_grpc.UploadAvatarSignRequest) (*user_grpc.UploadAvatarSignResponse, error) {
	resp, err := server.svc.UploadAvatarSign(ctx, req.UserID)
	if err != nil {
		return &user_grpc.UploadAvatarSignResponse{}, err
	}
	return &user_grpc.UploadAvatarSignResponse{Response: resp}, nil
}

func (server *UserServiceServer) UploadAvatarCallback(ctx context.Context, req *user_grpc.UploadAvatarCallbackRequest) (*user_grpc.UploadAvatarCallbackResponse, error) {
	if err := server.svc.UploadAvatarCallback(ctx, req.UserID, req.ObjectName); err != nil {
		return &user_grpc.UploadAvatarCallbackResponse{}, err
	}
	return &user_grpc.UploadAvatarCallbackResponse{}, nil
}

func (server *UserServiceServer) GetAvatarURL(ctx context.Context, req *user_grpc.GetAvatarURLRequest) (*user_grpc.GetAvatarURLResponse, error) {
	url, err := server.svc.GetAvatarURL(ctx, req.ObjectName)
	if err != nil {
		return &user_grpc.GetAvatarURLResponse{}, err
	}
	return &user_grpc.GetAvatarURLResponse{URL: url}, nil
}

func (server *UserServiceServer) HealthCheck(ctx context.Context, req *user_grpc.HealthCheckRequest) (*user_grpc.HealthCheckResponse, error) {
	return &user_grpc.HealthCheckResponse{}, nil
}
