package service

import (
	"github.com/yzletter/go-postery/dto/response"
	repository "github.com/yzletter/go-postery/repository/user"
)

type UserService struct {
	UserRepository *repository.GormUserRepository
}

func NewUserService(userRepository *repository.GormUserRepository) *UserService {
	return &UserService{UserRepository: userRepository}
}

func (svc *UserService) Register(name, password string) (dto.UserDTO, error) {
	user, err := svc.UserRepository.Create(name, password)
	return dto.ToUserDTO(user), err
}

func (svc *UserService) GetById(uid int) (bool, dto.UserDTO) {
	ok, user := svc.UserRepository.GetByID(uid)
	if !ok {
		return false, dto.UserDTO{}
	}

	return true, dto.ToUserDTO(user)
}

// GetByName 根据 name 查找用户
func (svc *UserService) GetByName(name string) dto.UserDTO {
	user, err := svc.UserRepository.GetByName(name)
	if err != nil {
		return dto.UserDTO{}
	}
	return dto.ToUserDTO(user)
}

func (svc *UserService) UpdatePassword(uid int, oldPass, newPass string) error {
	err := svc.UserRepository.UpdatePassword(uid, oldPass, newPass)
	return err
}

func (svc *UserService) Login(name, pass string) (bool, dto.UserDTO) {
	user, err := svc.UserRepository.GetByName(name)
	if err != nil || user.PassWord != pass {
		return false, dto.UserDTO{}
	}
	return true, dto.ToUserDTO(user)
}
