package service

import (
	"errors"

	"github.com/yzletter/go-postery/dto/request"
	"github.com/yzletter/go-postery/dto/response"
	"github.com/yzletter/go-postery/infra/snowflake"
	"github.com/yzletter/go-postery/model"
	"github.com/yzletter/go-postery/repository"
	"github.com/yzletter/go-postery/repository/dao"
)

var (
	ErrServerInternal = errors.New("注册失败, 请稍后重试")
	ErrNameDuplicated = errors.New("用户名重复")
)

type UserService struct {
	UserRepository repository.UserRepository
}

func NewUserService(userRepository repository.UserRepository) *UserService {
	return &UserService{
		UserRepository: userRepository,
	}
}

func (svc *UserService) Register(username, password string) (dto.UserBriefDTO, error) {
	var userDTO dto.UserBriefDTO

	u := &model.User{
		ID:           snowflake.NextID(),
		Username:     username,
		Email:        "",
		PasswordHash: password,
	}
	user, err := svc.UserRepository.Create(u)
	if err == nil {
		userDTO = dto.ToUserBriefDTO(*user)
		return userDTO, nil
	}

	if errors.Is(err, repository.ErrUniqueKeyConflict) { // 唯一键冲突
		return userDTO, ErrNameDuplicated
	} else if errors.Is(err, dao.ErrInternal) { // 数据库内部错误
		return userDTO, ErrServerInternal
	}
	return userDTO, ErrServerInternal
}

func (svc *UserService) GetBriefById(id int64) (bool, dto.UserBriefDTO) {
	user, err := svc.UserRepository.GetByID(id)
	if err != nil {
		return false, dto.UserBriefDTO{}
	}

	return true, dto.ToUserBriefDTO(*user)
}

// GetDetailById 根据 ID 查找用户的详细信息
func (svc *UserService) GetDetailById(id int64) (bool, dto.UserDetailDTO) {
	user, err := svc.UserRepository.GetByID(id)
	if err != nil {
		return false, dto.UserDetailDTO{}
	}

	return true, dto.ToUserDetailDTO(*user)
}

// GetBriefByName 根据 username 查找用户的简要信息
func (svc *UserService) GetBriefByName(username string) dto.UserBriefDTO {
	user, err := svc.UserRepository.GetByUsername(username)
	if err != nil {
		return dto.UserBriefDTO{}
	}
	return dto.ToUserBriefDTO(*user)
}

func (svc *UserService) UpdatePassword(id int64, oldPass, newPass string) error {
	err := svc.UserRepository.UpdatePasswordHash(id, newPass)
	return err
}

func (svc *UserService) UpdateProfile(id int64, req request.ModifyProfileRequest) error {
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
	if err := svc.UserRepository.UpdateProfile(id, updates); err == nil {
		return nil
	} else if errors.Is(err, dao.ErrRecordNotFound) {
		// 如果是用户 ID 错误, 直接返回该错误
		return err
	}
	return ErrServerInternal
}

func (svc *UserService) Login(username, pass string) (bool, dto.UserBriefDTO) {
	user, err := svc.UserRepository.GetByUsername(username)
	if err != nil || user.PasswordHash != pass {
		return false, dto.UserBriefDTO{}
	}
	return true, dto.ToUserBriefDTO(*user)
}

func (svc *UserService) CheckAdmin(id int64) (bool, error) {
	status, err := svc.UserRepository.GetStatus(id)
	if err != nil {
		return false, err
	}
	if status == 5 {
		return true, nil
	}
	return false, nil
}
