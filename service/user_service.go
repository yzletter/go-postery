package service

import (
	"context"

	userdto "github.com/yzletter/go-postery/dto/user"
	"github.com/yzletter/go-postery/errno"
	"github.com/yzletter/go-postery/model"
	"github.com/yzletter/go-postery/repository"
)

type userService struct {
	userRepo   repository.UserRepository // 依赖 UserRepository
	idGen      IDGenerator               // 用于生成 ID
	passHasher PasswordHasher            // 用于加密和比较密码
}

func NewUserService(userRepo repository.UserRepository, idGen IDGenerator, passHasher PasswordHasher) UserService {
	return &userService{
		userRepo:   userRepo,
		idGen:      idGen,
		passHasher: passHasher,
	}
}

// Register 注册用户
func (svc *userService) Register(ctx context.Context, username, email, password string) (userdto.BriefDTO, error) {
	var empty userdto.BriefDTO

	// 参数校验
	if username == "" || password == "" || email == "" {
		return empty, errno.ErrInvalidParam
	}
	if len(password) < 8 {
		return empty, errno.ErrPasswordWeak
	}

	// 对密码进行加密
	passwordHash, err := svc.passHasher.Hash(password)
	if err != nil {
		return empty, err
	}

	// 构造指针
	u := &model.User{
		ID:           svc.idGen.NextID(),
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
	}

	// 创建记录
	err = svc.userRepo.Create(ctx, u)
	if err != nil {
		return empty, toErrnoErr(err)
	}

	return userdto.ToBriefDTO(u), nil
}

// GetBriefById 根据 ID 查找用户的简要信息
func (svc *userService) GetBriefById(ctx context.Context, id int64) (userdto.BriefDTO, error) {
	var empty userdto.BriefDTO

	// 参数校验
	if id <= 0 {
		return empty, errno.ErrInvalidParam
	}

	// 获取用户
	user, err := svc.userRepo.GetByID(ctx, id)
	if err != nil {
		return empty, toErrnoErr(err)
	}

	// panic 兜底
	if user == nil {
		return empty, errno.ErrUserNotFound
	}

	return userdto.ToBriefDTO(user), nil
}

// GetDetailById 根据 ID 查找用户的详细信息
func (svc *userService) GetDetailById(ctx context.Context, id int64) (userdto.DetailDTO, error) {
	var empty userdto.DetailDTO

	// 参数校验
	if id <= 0 {
		return empty, errno.ErrInvalidParam
	}

	// 获取用户
	user, err := svc.userRepo.GetByID(ctx, id)
	if err != nil {
		return empty, toErrnoErr(err)
	}

	// panic 兜底
	if user == nil {
		return empty, errno.ErrUserNotFound
	}

	return userdto.ToDetailDTO(user), nil
}

// GetBriefByName 根据 username 查找用户的简要信息
func (svc *userService) GetBriefByName(ctx context.Context, username string) (userdto.BriefDTO, error) {
	var empty userdto.BriefDTO

	// 参数校验
	if len(username) <= 0 {
		return empty, errno.ErrInvalidParam
	}

	// 获取用户
	user, err := svc.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return empty, toErrnoErr(err)
	}

	// panic 兜底
	if user == nil {
		return empty, errno.ErrUserNotFound
	}
	return userdto.ToBriefDTO(user), nil
}

// UpdatePassword 更新密码
func (svc *userService) UpdatePassword(ctx context.Context, id int64, oldPass, newPass string) error {
	if id <= 0 || len(oldPass) <= 0 || len(newPass) <= 0 {
		return errno.ErrInvalidParam
	}

	if len(newPass) < 8 {
		return errno.ErrPasswordWeak
	}

	// todo 并发安全
	// 获取用户
	user, err := svc.userRepo.GetByID(ctx, id)
	if err != nil {
		return toErrnoErr(err)
	}
	if user == nil {
		return errno.ErrUserNotFound
	}

	// 判断旧密码是否正确
	err = svc.passHasher.Compare(user.PasswordHash, oldPass)
	if err != nil {
		return err
	}

	// 对新密码进行加密
	newPassHash, err := svc.passHasher.Hash(newPass)
	if err != nil {
		return err
	}

	// 改新密码
	err = svc.userRepo.UpdatePasswordHash(ctx, id, string(newPassHash))
	if err != nil {
		return toErrnoErr(err)
	}

	return nil
}

// UpdateProfile 修改个人资料
func (svc *userService) UpdateProfile(ctx context.Context, id int64, req userdto.ModifyProfileRequest) error {
	if id <= 0 {
		return errno.ErrInvalidParam
	}

	// 将 DTO 转为 Model, 主要是 Birthday 从 RFC3339 string 转为 Time.time
	modelReq := userdto.ModifyProfileRequestToModel(req)

	updates := map[string]any{
		"email":    modelReq.Email,
		"avatar":   modelReq.Avatar,
		"bio":      modelReq.Bio,
		"gender":   modelReq.Gender,
		"birthday": modelReq.BirthDay,
		"location": modelReq.Location,
		"country":  modelReq.Country,
	}

	if err := svc.userRepo.UpdateProfile(ctx, id, updates); err != nil {
		return toErrnoErr(err)
	}
	return nil
}

// Login 登录
func (svc *userService) Login(ctx context.Context, username, pass string) (userdto.BriefDTO, error) {
	var empty userdto.BriefDTO

	// 参数校验
	if username == "" || pass == "" {
		return empty, errno.ErrInvalidParam
	}

	// 获取用户
	user, err := svc.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return empty, toErrnoErr(err)
	}
	if user == nil {
		return empty, errno.ErrServerInternal
	}

	// 比较密码
	err = svc.passHasher.Compare(user.PasswordHash, pass)
	if err != nil {
		return empty, err
	}

	return userdto.ToBriefDTO(user), nil
}
