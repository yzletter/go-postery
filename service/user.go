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

func (service *UserService) GetUserById(uid int) *model.User {
	user := service.UserRepository.GetByID(uid)
	return user
}
