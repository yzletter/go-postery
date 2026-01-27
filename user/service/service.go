package service

import (
	"context"
	"errors"

	user_grpc "github.com/yzletter/go-postery/api/proto/user/v1"
	"github.com/yzletter/go-postery/errno"
	"github.com/yzletter/go-postery/user/dto"
	"github.com/yzletter/go-postery/user/repository"
	"github.com/yzletter/go-postery/user/service/ports"
)

type userService struct {
	repository repository.UserRepository // 依赖 UserRepository
	idGen      ports.IDGenerator         // 用于生成 ID
	passHasher ports.PasswordHasher      // 用于加密和比较密码
	user_grpc.UnimplementedUserServiceServer
}

func NewUserService(repository repository.UserRepository, idGen ports.IDGenerator, passHasher ports.PasswordHasher) UserService {
	return &userService{
		repository:                     repository,
		idGen:                          idGen,
		passHasher:                     passHasher,
		UnimplementedUserServiceServer: user_grpc.UnimplementedUserServiceServer{},
	}
}

func (svc *userService) GetProfileById(ctx context.Context, req *user_grpc.GetProfileByIdRequest) (*user_grpc.GetProfileByIdResponse, error) {
	empty := new(user_grpc.GetProfileByIdResponse)

	if req.ID <= 0 {
		return empty, errno.ErrInvalidParam
	}

	// 获取用户
	profile, err := svc.repository.GetProfileByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return empty, errno.ErrUserNotFound
		}
		return empty, errno.ErrServerInternal
	}

	if profile == nil {
		return empty, errno.ErrUserNotFound
	}

	return dto.ToGetProfileByIdResponse(profile), nil
}

func (svc *userService) UpdateProfile(ctx context.Context, req *user_grpc.UpdateProfileRequest) (*user_grpc.UpdateProfileResponse, error) {
	empty := new(user_grpc.UpdateProfileResponse)

	if req.ID <= 0 {
		return empty, errno.ErrInvalidParam
	}

	// 将 Request 转为 Model, 主要是 Birthday 从 RFC3339 string 转为 Time.time
	profile := dto.UpdateProfileRequestToModel(req)

	updates := map[string]any{
		"nickname": profile.Nickname,
		"avatar":   profile.Avatar,
		"bio":      profile.Bio,
		"gender":   profile.Gender,
		"birthday": profile.Birthday,
		"location": profile.Location,
		"country":  profile.Country,
	}

	if err := svc.repository.UpdateProfile(ctx, req.ID, updates); err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return empty, errno.ErrUserNotFound
		}
		return empty, errno.ErrServerInternal
	}

	return &user_grpc.UpdateProfileResponse{}, nil
}

func (svc *userService) Top(ctx context.Context, req *user_grpc.TopRequest) (*user_grpc.TopResponse, error) {
	empty := new(user_grpc.TopResponse)
	profiles, scores, err := svc.repository.Top(ctx)
	if err != nil {
		return empty, errno.ErrServerInternal
	}

	var topUsers []*user_grpc.TopUser
	for idx, profile := range profiles {
		topUsers = append(topUsers, dto.ToTopUser(profile, scores[idx]))
	}

	return &user_grpc.TopResponse{TopUsers: topUsers}, nil
}
