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
	uid, err := svc.UserRepository.Create(name, password)
	back := dto.UserDTO{
		Id:   uid,
		Name: name,
	}
	return back, err
}

func (svc *UserService) GetById(uid int) (bool, dto.UserDTO) {
	ok, user := svc.UserRepository.GetByID(uid)
	if !ok {
		return false, dto.UserDTO{}
	}

	userDTO := dto.UserDTO{
		Id:   user.Id,
		Name: user.Name,
	}
	return true, userDTO
}

// GetByName 根据 name 查找用户
func (svc *UserService) GetByName(name string) dto.UserDTO {
	user := svc.UserRepository.GetByName(name)
	if user == nil {
		return dto.UserDTO{}
	}

	userDTO := dto.UserDTO{
		Id:   user.Id,
		Name: user.Name,
	}
	return userDTO
}

func (svc *UserService) UpdatePassword(uid int, oldPass, newPass string) error {
	err := svc.UserRepository.UpdatePassword(uid, oldPass, newPass)
	return err
}

func (svc *UserService) Login(name, pass string) (bool, dto.UserDTO) {
	user := svc.UserRepository.GetByName(name)
	if user == nil || user.PassWord != pass {
		return false, dto.UserDTO{}
	}

	userDTO := dto.UserDTO{
		Id:   user.Id,
		Name: user.Name,
	}
	return true, userDTO
}
