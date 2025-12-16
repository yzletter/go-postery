package repository

import (
	"context"

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

func (repo *userRepository) Create(ctx context.Context, user *model.User) error {
	err := repo.dao.Create(ctx, user)
	if err != nil {
		return toRepoErr(err)
	}

	// todo 写 Cache

	return nil
}

func (repo *userRepository) Delete(ctx context.Context, id int64) error {
	err := repo.dao.Delete(ctx, id)
	if err != nil {
		return toRepoErr(err)
	}

	// todo 删 Cache

	return nil
}

func (repo *userRepository) GetPasswordHash(ctx context.Context, id int64) (string, error) {
	passwordHash, err := repo.dao.GetPasswordHash(ctx, id)
	if err != nil {
		return "", toRepoErr(err)
	}
	return passwordHash, nil
}

func (repo *userRepository) GetStatus(ctx context.Context, id int64) (int, error) {
	// todo 查 Cache

	status, err := repo.dao.GetStatus(ctx, id)
	if err != nil {
		return 0, toRepoErr(err)
	}
	return status, nil
}

func (repo *userRepository) GetByID(ctx context.Context, id int64) (*model.User, error) {
	// todo 查 Cache

	user, err := repo.dao.GetByID(ctx, id)
	if err != nil {
		return nil, toRepoErr(err)
	}
	return user, nil
}

func (repo *userRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	// todo 查 Cache

	user, err := repo.dao.GetByUsername(ctx, username)
	if err != nil {
		return nil, toRepoErr(err)
	}
	return user, nil
}

func (repo *userRepository) UpdatePasswordHash(ctx context.Context, id int64, newHash string) error {
	err := repo.dao.UpdatePasswordHash(ctx, id, newHash)
	if err != nil {
		return toRepoErr(err)
	}

	return nil
}

func (repo *userRepository) UpdateProfile(ctx context.Context, id int64, updates map[string]any) error {
	err := repo.dao.UpdateProfile(ctx, id, updates)
	if err != nil {
		return toRepoErr(err)
	}

	// todo 更新 Cache

	return nil
}
