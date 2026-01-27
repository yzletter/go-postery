package service

import (
	"context"

	user_grpc "github.com/yzletter/go-postery/api/proto/user/v1"
)

type UserService interface {
	GetProfileById(ctx context.Context, req *user_grpc.GetProfileByIdRequest) (*user_grpc.GetProfileByIdResponse, error) // 根据用户 ID 获取用户资料
	UpdateProfile(ctx context.Context, req *user_grpc.UpdateProfileRequest) (*user_grpc.UpdateProfileResponse, error)    // 更新用户资料
	Top(ctx context.Context, req *user_grpc.TopRequest) (*user_grpc.TopResponse, error)                                  // 返回推荐用户
	user_grpc.UnsafeUserServiceServer
}
