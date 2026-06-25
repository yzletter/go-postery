package grpc

import (
	"context"
	"time"

	user_grpc "github.com/yzletter/go-postery/api/proto/user/v1"
	"github.com/yzletter/go-postery/backend/grpc/errs"
	"github.com/yzletter/go-postery/backend/micro/user/grpc/dto"
	"github.com/yzletter/go-postery/backend/micro/user/service"
)

// UserServiceServer 实现用户 gRPC 服务
type UserServiceServer struct {
	svc service.UserService
	user_grpc.UnimplementedUserServiceServer
}

// NewUserServiceServer 构造函数
func NewUserServiceServer(svc service.UserService) *UserServiceServer {
	return &UserServiceServer{
		svc: svc,
	}
}

// GetProfileById 根据 ID 获取用户资料
func (server *UserServiceServer) GetProfileById(ctx context.Context, req *user_grpc.GetProfileByIdRequest) (*user_grpc.Profile, error) {
	if req == nil || req.ID <= 0 {
		return &user_grpc.Profile{}, errs.ErrInvalidArgument
	}

	profile, err := server.svc.GetProfile(ctx, req.ID)
	if err != nil {
		return &user_grpc.Profile{}, err
	}
	return dto.ToProfile(profile), nil
}

// UpdateProfile 更新用户资料
func (server *UserServiceServer) UpdateProfile(ctx context.Context, req *user_grpc.UpdateProfileRequest) (*user_grpc.UpdateProfileResponse, error) {
	if req == nil || req.UserID <= 0 {
		return &user_grpc.UpdateProfileResponse{}, errs.ErrInvalidArgument
	}

	updates, err := dto.UpdateProfileRequestToMap(req)
	if err != nil {
		return &user_grpc.UpdateProfileResponse{}, errs.ErrInvalidArgument
	}

	if err := server.svc.UpdateProfile(ctx, req.UserID, updates); err != nil {
		return &user_grpc.UpdateProfileResponse{}, err
	}
	return &user_grpc.UpdateProfileResponse{}, nil
}

// Top 获取用户排行榜
func (server *UserServiceServer) Top(ctx context.Context, req *user_grpc.TopRequest) (*user_grpc.TopResponse, error) {
	if req == nil {
		return &user_grpc.TopResponse{}, errs.ErrInvalidArgument
	}

	profiles, err := server.svc.Top(ctx)
	if err != nil {
		return &user_grpc.TopResponse{}, err
	}

	profileTops := make([]*user_grpc.ProfileTop, 0, len(profiles))
	for _, profile := range profiles {
		profileTops = append(profileTops, dto.ToProfileTop(profile))
	}

	return &user_grpc.TopResponse{ProfileTops: profileTops}, nil
}

// GetIDAfterTime 根据时间获取之后创建的用户 ID
func (server *UserServiceServer) GetIDAfterTime(ctx context.Context, req *user_grpc.GetIDAfterTimeRequest) (*user_grpc.UserIDs, error) {
	if req == nil || req.TimeAfter == "" {
		return &user_grpc.UserIDs{}, errs.ErrInvalidArgument
	}

	timeAt, err := time.Parse(time.RFC3339, req.TimeAfter)
	if err != nil {
		return &user_grpc.UserIDs{}, errs.ErrInvalidArgument
	}

	ids, err := server.svc.GetIDAfterTime(ctx, timeAt)
	if err != nil {
		return &user_grpc.UserIDs{}, err
	}
	return &user_grpc.UserIDs{IDs: ids}, nil
}

// ListFollowersByPage 按页获取粉丝
func (server *UserServiceServer) ListFollowersByPage(ctx context.Context, req *user_grpc.ListFollowRequest) (*user_grpc.ListFollowResponse, error) {
	if req == nil || req.UserID <= 0 || req.PageNo == 0 || req.PageSize == 0 || req.PageSize > 100 {
		return &user_grpc.ListFollowResponse{}, errs.ErrInvalidArgument
	}

	total, profiles, err := server.svc.ListFollowers(ctx, req.UserID, int(req.PageNo), int(req.PageSize))
	if err != nil {
		return &user_grpc.ListFollowResponse{}, err
	}

	profileBriefs := make([]*user_grpc.ProfileBrief, 0, len(profiles))
	for _, profile := range profiles {
		profileBriefs = append(profileBriefs, dto.ToProfileBrief(profile))
	}

	return &user_grpc.ListFollowResponse{Count: uint64(total), ProfileBriefs: profileBriefs}, nil
}

// ListFolloweesByPage 按页获取关注的人
func (server *UserServiceServer) ListFolloweesByPage(ctx context.Context, req *user_grpc.ListFollowRequest) (*user_grpc.ListFollowResponse, error) {
	if req == nil || req.UserID <= 0 || req.PageNo == 0 || req.PageSize == 0 || req.PageSize > 100 {
		return &user_grpc.ListFollowResponse{}, errs.ErrInvalidArgument
	}

	total, profiles, err := server.svc.ListFollowees(ctx, req.UserID, int(req.PageNo), int(req.PageSize))
	if err != nil {
		return &user_grpc.ListFollowResponse{}, err
	}

	profileBriefs := make([]*user_grpc.ProfileBrief, 0, len(profiles))
	for _, profile := range profiles {
		profileBriefs = append(profileBriefs, dto.ToProfileBrief(profile))
	}

	return &user_grpc.ListFollowResponse{Count: uint64(total), ProfileBriefs: profileBriefs}, nil
}

// UploadAvatarSign 获取上传头像 OSS 签名
func (server *UserServiceServer) UploadAvatarSign(ctx context.Context, req *user_grpc.UploadAvatarSignRequest) (*user_grpc.UploadAvatarSignResponse, error) {
	if req == nil || req.UserID <= 0 {
		return &user_grpc.UploadAvatarSignResponse{}, errs.ErrInvalidArgument
	}

	resp, err := server.svc.UploadAvatarSign(ctx, req.UserID)
	if err != nil {
		return &user_grpc.UploadAvatarSignResponse{}, err
	}
	return &user_grpc.UploadAvatarSignResponse{Response: resp}, nil
}

// UploadAvatarCallback 处理头像上传回调
func (server *UserServiceServer) UploadAvatarCallback(ctx context.Context, req *user_grpc.UploadAvatarCallbackRequest) (*user_grpc.UploadAvatarCallbackResponse, error) {
	if req == nil || req.UserID <= 0 || req.ObjectName == "" {
		return &user_grpc.UploadAvatarCallbackResponse{}, errs.ErrInvalidArgument
	}

	if err := server.svc.UploadAvatarCallback(ctx, req.UserID, req.ObjectName); err != nil {
		return &user_grpc.UploadAvatarCallbackResponse{}, err
	}
	return &user_grpc.UploadAvatarCallbackResponse{}, nil
}

// GetAvatarURL 获取头像访问预签名 URL
func (server *UserServiceServer) GetAvatarURL(ctx context.Context, req *user_grpc.GetAvatarURLRequest) (*user_grpc.GetAvatarURLResponse, error) {
	if req == nil || req.ObjectName == "" {
		return &user_grpc.GetAvatarURLResponse{}, errs.ErrInvalidArgument
	}

	url, err := server.svc.GetAvatarURL(ctx, req.ObjectName)
	if err != nil {
		return &user_grpc.GetAvatarURLResponse{}, err
	}
	return &user_grpc.GetAvatarURLResponse{URL: url}, nil
}

// HealthCheck 健康检查
func (server *UserServiceServer) HealthCheck(ctx context.Context, req *user_grpc.HealthCheckRequest) (*user_grpc.HealthCheckResponse, error) {
	return &user_grpc.HealthCheckResponse{}, nil
}
