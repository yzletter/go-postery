package service

import (
	"context"
	"errors"

	userdto "github.com/yzletter/go-postery/dto/user"
	"github.com/yzletter/go-postery/errno"
	"github.com/yzletter/go-postery/repository"
	"github.com/yzletter/go-postery/service/ports"
)

type userService struct {
	userRepo   repository.UserRepository // 依赖 UserRepository
	idGen      ports.IDGenerator         // 用于生成 ID
	passHasher ports.PasswordHasher      // 用于加密和比较密码
}

func NewUserService(userRepo repository.UserRepository, idGen ports.IDGenerator, passHasher ports.PasswordHasher) UserService {
	return &userService{
		userRepo:   userRepo,
		idGen:      idGen,
		passHasher: passHasher,
	}
}

// GetProfileById 根据 ID 查找用户的资料
func (svc *userService) GetProfileById(ctx context.Context, id int64) (userdto.DetailDTO, error) {
	var empty userdto.DetailDTO

	// 参数校验
	if id <= 0 {
		return empty, errno.ErrInvalidParam
	}

	// 获取用户
	userProfile, err := svc.userRepo.GetProfileByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return empty, errno.ErrUserNotFound
		}
		return empty, errno.ErrServerInternal
	}

	// panic 兜底
	if userProfile == nil {
		return empty, errno.ErrUserNotFound
	}

	return userdto.ToDetailDTO(userProfile), nil
}

// UpdateProfile 修改个人资料
func (svc *userService) UpdateProfile(ctx context.Context, id int64, req userdto.ModifyProfileRequest) error {
	if id <= 0 {
		return errno.ErrInvalidParam
	}

	// 将 DTO 转为 Model, 主要是 Birthday 从 RFC3339 string 转为 Time.time
	userProfile := userdto.ModifyProfileRequestToModel(req)

	updates := map[string]any{
		"nickname": userProfile.Nickname,
		"avatar":   userProfile.Avatar,
		"bio":      userProfile.Bio,
		"gender":   userProfile.Gender,
		"birthday": userProfile.BirthDay,
		"location": userProfile.Location,
		"country":  userProfile.Country,
	}

	if err := svc.userRepo.UpdateProfile(ctx, id, updates); err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return errno.ErrUserNotFound
		}
		return errno.ErrServerInternal
	}
	return nil
}

// Top 返回热门推荐用户
func (svc *userService) Top(ctx context.Context) ([]userdto.TopDTO, error) {
	var empty []userdto.TopDTO
	userProfiles, scores, err := svc.userRepo.Top(ctx)
	if err != nil {
		return empty, errno.ErrServerInternal
	}

	var userDTOs []userdto.TopDTO
	for idx, userProfile := range userProfiles {
		userDTOs = append(userDTOs, userdto.ToTopDTO(userProfile, scores[idx]))
	}

	return userDTOs, nil
}
