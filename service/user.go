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

func (svc *UserService) Register(name, password string) (int, error) {
	uid, err := svc.UserRepository.Create(name, password)
	return uid, err
}

func (svc *UserService) GetById(uid int) *model.User {
	user := svc.UserRepository.GetByID(uid)
	return user
}

// GetByName 根据 name 查找用户
func (svc *UserService) GetByName(name string) *model.User {
	user := svc.UserRepository.GetByName(name)
	return user
}

func (svc *UserService) UpdatePassword(uid int, oldPass, newPass string) error {
	err := svc.UserRepository.UpdatePassword(uid, oldPass, newPass)
	return err
}
