package repository

import (
	"github.com/yzletter/go-postery/model"
	"github.com/yzletter/go-postery/repository/cache"
	"github.com/yzletter/go-postery/repository/dao"
)

// todo 错误映射

type userRepository struct {
	dao   dao.UserDAO
	cache cache.UserCache
}

func NewUserRepository(userDAO dao.UserDAO, userCache cache.UserCache) UserRepository {
	return &userRepository{dao: userDAO, cache: userCache}
}

func (repo *userRepository) Create(user *model.User) (*model.User, error) {
	u, err := repo.dao.Create(user)
	if err != nil {
		return nil, err
	}

	// todo 写 Cache

	return u, nil
}

func (repo *userRepository) Delete(id int64) error {
	err := repo.dao.Delete(id)
	if err != nil {
		return err
	}

	// todo 删 Cache

	return nil
}

func (repo *userRepository) GetPasswordHash(id int64) (string, error) {
	passwordHash, err := repo.dao.GetPasswordHash(id)
	if err != nil {
		return "", err
	}
	return passwordHash, nil
}

func (repo *userRepository) GetStatus(id int64) (uint8, error) {
	// todo 查 Cache

	status, err := repo.dao.GetStatus(id)
	if err != nil {
		return 0, err
	}
	return status, nil
}

func (repo *userRepository) GetByID(id int64) (*model.User, error) {
	// todo 查 Cache

	user, err := repo.dao.GetByID(id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (repo *userRepository) GetByUsername(username string) (*model.User, error) {
	// todo 查 Cache

	user, err := repo.dao.GetByUsername(username)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (repo *userRepository) UpdatePasswordHash(id int64, newHash string) error {
	return repo.dao.UpdatePasswordHash(id, newHash)
}

func (repo *userRepository) UpdateProfile(id int64, updates map[string]any) error {
	err := repo.dao.UpdateProfile(id, updates)
	if err != nil {
		return err
	}

	// todo 更新 Cache

	return nil
}
