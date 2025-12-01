package service

import (
	"github.com/yzletter/go-postery/model"
	repository "github.com/yzletter/go-postery/repository/user"
)

type UserService struct {
	UserRepository *repository.GormUserRepository
}

func NewUserService(userRepository *repository.GormUserRepository) *UserService {
	return &UserService{UserRepository: userRepository}
}

func (service *UserService) Register(name, password string) (int, error) {
	uid, err := service.UserRepository.Create(name, password)
	return uid, err
}

func (service *UserService) GetById(uid int) *model.User {
	user := service.UserRepository.GetByID(uid)
	return user
}

// GetByName 根据 name 查找用户
func (service *UserService) GetByName(name string) *model.User {
	user := service.UserRepository.GetByName(name)
	return user
}

func (service *UserService) UpdatePassword(uid int, oldPass, newPass string) (bool, error) {
	ok, err := service.UserRepository.UpdatePassword(uid, oldPass, newPass)
	return ok, err
}
