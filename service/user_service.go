package service

import (
	"context"
	"errors"

	"github.com/yzletter/go-postery/dto/request"
	"github.com/yzletter/go-postery/dto/response"
	"github.com/yzletter/go-postery/infra/snowflake"
	"github.com/yzletter/go-postery/model"
	"github.com/yzletter/go-postery/repository"
	"github.com/yzletter/go-postery/repository/dao"
	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	UserRepository repository.UserRepository
}

func NewUserService(userRepository repository.UserRepository) UserService {
	return &userService{
		UserRepository: userRepository,
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
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
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
		if errors.Is(err, dao.ErrUniqueKey) { // 唯一键冲突
			return empty, ErrDuplicated
		} else if errors.Is(err, dao.ErrInternal) { // 数据库内部错误
			return empty, ErrServerInternal
		}
		return empty, ErrServerInternal
	}

	return dto.ToUserBriefDTO(user), nil
}

// GetBriefById 根据 ID 查找用户的简要信息
func (svc *userService) GetBriefById(ctx context.Context, id int64) (bool, dto.UserBriefDTO) {
	user, err := svc.UserRepository.GetByID(ctx, id)
	if err != nil {
		return false, dto.UserBriefDTO{}
	}

	return true, dto.ToUserBriefDTO(*user)
}

// GetDetailById 根据 ID 查找用户的详细信息
func (svc *userService) GetDetailById(ctx context.Context, id int64) (bool, dto.UserDetailDTO) {
	user, err := svc.UserRepository.GetByID(ctx, id)
	if err != nil {
		return false, dto.UserDetailDTO{}
	}

	return true, dto.ToUserDetailDTO(*user)
}

// GetBriefByName 根据 username 查找用户的简要信息
func (svc *userService) GetBriefByName(ctx context.Context, username string) dto.UserBriefDTO {
	user, err := svc.UserRepository.GetByUsername(ctx, username)
	if err != nil {
		return dto.UserBriefDTO{}
	}
	return dto.ToUserBriefDTO(*user)
}

func (svc *userService) UpdatePassword(ctx context.Context, id int64, oldPass, newPass string) error {
	err := svc.UserRepository.UpdatePasswordHash(ctx, id, newPass)
	return err
}

func (svc *userService) UpdateProfile(ctx context.Context, id int64, req request.ModifyProfileRequest) error {
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
	if err := svc.UserRepository.UpdateProfile(ctx, id, updates); err == nil {
		return nil
	} else if errors.Is(err, dao.ErrRecordNotFound) {
		// 如果是用户 ID 错误, 直接返回该错误
		return err
	}
	return ErrServerInternal
}

func (svc *userService) Login(ctx context.Context, username, pass string) (bool, dto.UserBriefDTO) {
	user, err := svc.UserRepository.GetByUsername(ctx, username)
	if err != nil || user.PasswordHash != pass {
		return false, dto.UserBriefDTO{}
	}
	return true, dto.ToUserBriefDTO(*user)
}

func (svc *userService) CheckAdmin(ctx context.Context, id int64) (bool, error) {
	status, err := svc.UserRepository.GetStatus(ctx, id)
	if err != nil {
		return false, err
	}
	if status == 5 {
		return true, nil
	}
	return false, nil
}
