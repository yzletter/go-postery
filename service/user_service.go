package service

import (
	"context"

	"github.com/yzletter/go-postery/dto/request"
	"github.com/yzletter/go-postery/dto/response"
	"github.com/yzletter/go-postery/infra/snowflake"
	"github.com/yzletter/go-postery/model"
	"github.com/yzletter/go-postery/repository"
)

type userService struct {
	UserRepository repository.UserRepository
	PasswordHasher PasswordHasher // 用于加密和比较密码
}

func NewUserService(userRepository repository.UserRepository, passwordHasher PasswordHasher) UserService {
	return &userService{
		UserRepository: userRepository,
		PasswordHasher: passwordHasher,
	}
}

// Register 注册用户
func (svc *userService) Register(ctx context.Context, username, email, password string) (dto.UserBriefDTO, error) {
	var empty dto.UserBriefDTO

	// 参数校验
	if username == "" || password == "" || email == "" {
		return empty, ErrInvalidParam
	}
	if len(password) < 8 {
		return empty, ErrPasswordWeak
	}

	// 对密码进行加密
	passwordHash, err := svc.PasswordHasher.Hash(password)
	if err != nil {
		return empty, ErrEncryptFailed
	}

	// 构造指针
	user := &model.User{
		ID:           snowflake.NextID(),
		Username:     username,
		Email:        email,
		PasswordHash: string(passwordHash),
	}

	// 创建记录
	err = svc.UserRepository.Create(ctx, user)
	if err != nil {
		return empty, toServiceErr(err)
	}

	return dto.ToUserBriefDTO(user), nil
}

// GetBriefById 根据 ID 查找用户的简要信息
func (svc *userService) GetBriefById(ctx context.Context, id int64) (dto.UserBriefDTO, error) {
	var empty dto.UserBriefDTO

	// 参数校验
	if id <= 0 {
		return empty, ErrInvalidParam
	}

	user, err := svc.UserRepository.GetByID(ctx, id)
	if err != nil {
		return empty, toServiceErr(err)
	}

	// panic 兜底
	if user == nil {
		return empty, ErrNotFound
	}

	return dto.ToUserBriefDTO(user), nil
}

// GetDetailById 根据 ID 查找用户的详细信息
func (svc *userService) GetDetailById(ctx context.Context, id int64) (dto.UserDetailDTO, error) {
	var empty dto.UserDetailDTO

	// 参数校验
	if id <= 0 {
		return empty, ErrInvalidParam
	}

	user, err := svc.UserRepository.GetByID(ctx, id)
	if err != nil {
		return empty, toServiceErr(err)
	}

	// panic 兜底
	if user == nil {
		return empty, ErrNotFound
	}

	return dto.ToUserDetailDTO(user), nil
}

// GetBriefByName 根据 username 查找用户的简要信息
func (svc *userService) GetBriefByName(ctx context.Context, username string) (dto.UserBriefDTO, error) {
	var empty dto.UserBriefDTO

	// 参数校验
	if len(username) <= 0 {
		return empty, ErrInvalidParam
	}

	user, err := svc.UserRepository.GetByUsername(ctx, username)
	if err != nil {
		return empty, toServiceErr(err)
	}

	// panic 兜底
	if user == nil {
		return empty, ErrNotFound
	}
	return dto.ToUserBriefDTO(user), nil
}

// UpdatePassword 更新密码
func (svc *userService) UpdatePassword(ctx context.Context, id int64, oldPass, newPass string) error {
	if id <= 0 || len(oldPass) <= 0 || len(newPass) <= 0 {
		return ErrInvalidParam
	}

	if len(newPass) < 8 {
		return ErrPasswordWeak
	}

	// todo 并发安全
	user, err := svc.UserRepository.GetByID(ctx, id)
	if err != nil {
		return toServiceErr(err)
	}
	if user == nil {
		return ErrNotFound
	}

	// 判断旧密码是否正确
	err = svc.PasswordHasher.Compare(user.PasswordHash, oldPass)
	if err != nil {
		return ErrEncryptFailed
	}

	// 对新密码进行加密
	newPassHash, err := svc.PasswordHasher.Hash(newPass)
	if err != nil {
		return ErrEncryptFailed
	}

	// 改新密码
	err = svc.UserRepository.UpdatePasswordHash(ctx, id, string(newPassHash))
	if err != nil {
		return toServiceErr(err)
	}

	return nil
}

// UpdateProfile 修改个人资料
func (svc *userService) UpdateProfile(ctx context.Context, id int64, req request.ModifyProfileRequest) error {
	if id <= 0 {
		return ErrInvalidParam
	}

	// 将 DTO 转为 Model, 主要是 Birthday 从 RFC3339 string 转为 Time.time
	modelReq := request.ModifyProfileRequestToModel(req)

	updates := map[string]any{
		"email":    modelReq.Email,
		"avatar":   modelReq.Avatar,
		"bio":      modelReq.Bio,
		"gender":   modelReq.Gender,
		"birthday": modelReq.BirthDay,
		"location": modelReq.Location,
		"country":  modelReq.Country,
	}

	if err := svc.UserRepository.UpdateProfile(ctx, id, updates); err != nil {
		return toServiceErr(err)
	}
	return nil
}

// Login 登录
func (svc *userService) Login(ctx context.Context, username, pass string) (dto.UserBriefDTO, error) {
	var empty dto.UserBriefDTO

	// 参数校验
	if username == "" || pass == "" {
		return empty, ErrInvalidParam
	}

	user, err := svc.UserRepository.GetByUsername(ctx, username)
	if err != nil {
		return empty, toServiceErr(err)
	}
	if user == nil {
		return empty, ErrServerInternal
	}

	err = svc.PasswordHasher.Compare(user.PasswordHash, pass)
	if err != nil {
		return empty, ErrEncryptFailed
	}

	return dto.ToUserBriefDTO(user), nil
}
