package service

import (
	"errors"

	"github.com/yzletter/go-postery/dto/request"
	"github.com/yzletter/go-postery/dto/response"
	repository "github.com/yzletter/go-postery/repository/user"
)

var (
	ErrServerInternal = errors.New("注册失败, 请稍后重试")
	ErrNameDuplicated = errors.New("用户名重复")
)

type UserService struct {
	UserDBRepo    *repository.UserDBRepository
	UserCacheRepo *repository.UserCacheRepository
}

func NewUserService(userDBRepository *repository.UserDBRepository, userCacheRepository *repository.UserCacheRepository) *UserService {
	return &UserService{
		UserDBRepo:    userDBRepository,
		UserCacheRepo: userCacheRepository,
	}
}

func (svc *UserService) Register(name, password, ip string) (dto.UserBriefDTO, error) {
	var userDTO dto.UserBriefDTO
	user, err := svc.UserDBRepo.Create(name, password, ip)
	if err == nil {
		userDTO = dto.ToUserBriefDTO(user)
		return userDTO, nil
	}

	if errors.Is(err, repository.ErrUniqueKeyConflict) { // 唯一键冲突
		return userDTO, ErrNameDuplicated
	} else if errors.Is(err, repository.ErrMySQLInternal) { // 数据库内部错误
		return userDTO, ErrServerInternal
	}
	return userDTO, ErrServerInternal
}

func (svc *UserService) GetBriefById(uid int) (bool, dto.UserBriefDTO) {
	ok, user := svc.UserDBRepo.GetByID(uid)
	if !ok {
		return false, dto.UserBriefDTO{}
	}

	return true, dto.ToUserBriefDTO(user)
}

// GetDetailById 根据 Id 查找用户的详细信息
func (svc *UserService) GetDetailById(uid int) (bool, dto.UserDetailDTO) {
	ok, user := svc.UserDBRepo.GetByID(uid)
	if !ok {
		return false, dto.UserDetailDTO{}
	}

	return true, dto.ToUserDetailDTO(user)
}

// GetBriefByName 根据 name 查找用户的简要信息
func (svc *UserService) GetBriefByName(name string) dto.UserBriefDTO {
	user, err := svc.UserDBRepo.GetByName(name)
	if err != nil {
		return dto.UserBriefDTO{}
	}
	return dto.ToUserBriefDTO(user)
}

func (svc *UserService) UpdatePassword(uid int, oldPass, newPass string) error {
	err := svc.UserDBRepo.UpdatePassword(uid, oldPass, newPass)
	return err
}

func (svc *UserService) UpdateProfile(uid int, req request.ModifyProfileRequest) error {
	// 将 DTO 转为 Model, 主要是 Birthday 从 RFC3339 string 转为 Time.time
	modelReq := request.ModifyProfileRequestToModel(req)

	if err := svc.UserDBRepo.UpdateProfile(uid, modelReq); err == nil {
		return nil
	} else if errors.Is(err, repository.ErrUidInvalid) {
		// 如果是用户 ID 错误, 直接返回该错误
		return err
	}
	return ErrServerInternal
}

func (svc *UserService) Login(name, pass string) (bool, dto.UserBriefDTO) {
	user, err := svc.UserDBRepo.GetByName(name)
	if err != nil || user.PassWord != pass {
		return false, dto.UserBriefDTO{}
	}
	return true, dto.ToUserBriefDTO(user)
}
