package service

import (
	"context"

	user_grpc "github.com/yzletter/go-postery/api/proto/user/v1"
)

type UserService interface {
	GetProfileById(ctx context.Context, req *user_grpc.GetProfileByIdRequest) (*user_grpc.UserDetail, error)          // 根据用户 ID 获取用户资料
	UpdateProfile(ctx context.Context, req *user_grpc.UpdateProfileRequest) (*user_grpc.UpdateProfileResponse, error) // 更新用户资料
	Top(ctx context.Context, req *user_grpc.TopRequest) (*user_grpc.TopResponse, error)                               // 返回推荐用户
	Follow(ctx context.Context, req *user_grpc.FollowCommonRequest) (*user_grpc.FollowEmptyResponse, error)
	UnFollow(ctx context.Context, req *user_grpc.FollowCommonRequest) (*user_grpc.FollowEmptyResponse, error)
	IfFollow(ctx context.Context, req *user_grpc.FollowCommonRequest) (*user_grpc.IfFollowResponse, error)
	ListFollowersByPage(ctx context.Context, req *user_grpc.ListFollowRequest) (*user_grpc.ListFollowResponse, error) // 按页查找粉丝
	ListFolloweesByPage(ctx context.Context, req *user_grpc.ListFollowRequest) (*user_grpc.ListFollowResponse, error) // 按页查找关注的人
	user_grpc.UnsafeUserServiceServer
}
