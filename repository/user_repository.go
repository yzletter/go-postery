package repository

import (
	"context"
	"log/slog"

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
		return toRepositoryErr(err)
	}

	// todo 写 Cache

	return nil
}

func (repo *userRepository) Delete(ctx context.Context, id int64) error {
	err := repo.dao.Delete(ctx, id)
	if err != nil {
		return toRepositoryErr(err)
	}

	// todo 删 Cache

	return nil
}

func (repo *userRepository) GetPasswordHash(ctx context.Context, id int64) (string, error) {
	passwordHash, err := repo.dao.GetPasswordHash(ctx, id)
	if err != nil {
		return "", toRepositoryErr(err)
	}
	return passwordHash, nil
}

func (repo *userRepository) GetStatus(ctx context.Context, id int64) (int, error) {
	// todo 查 Cache

	status, err := repo.dao.GetStatus(ctx, id)
	if err != nil {
		return 0, toRepositoryErr(err)
	}
	return status, nil
}

func (repo *userRepository) GetByID(ctx context.Context, id int64) (*model.User, error) {
	// todo 查 Cache

	user, err := repo.dao.GetByID(ctx, id)
	if err != nil {
		return nil, toRepositoryErr(err)
	}
	return user, nil
}

func (repo *userRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	// todo 查 Cache

	user, err := repo.dao.GetByUsername(ctx, username)
	if err != nil {
		return nil, toRepositoryErr(err)
	}
	return user, nil
}

func (repo *userRepository) UpdatePasswordHash(ctx context.Context, id int64, newHash string) error {
	err := repo.dao.UpdatePasswordHash(ctx, id, newHash)
	if err != nil {
		return toRepositoryErr(err)
	}

	return nil
}

func (repo *userRepository) UpdateProfile(ctx context.Context, id int64, updates map[string]any) error {
	err := repo.dao.UpdateProfile(ctx, id, updates)
	if err != nil {
		return toRepositoryErr(err)
	}

	// todo 更新 Cache

	return nil
}

func (repo *userRepository) Top(ctx context.Context) ([]*model.User, []float64, error) {
	ids, scores, err := repo.cache.Top(ctx)
	if err != nil {
		return nil, nil, toRepositoryErr(err)
	}

	var users []*model.User
	for _, id := range ids {
		user, err := repo.dao.GetByID(ctx, id)
		if err != nil {
			user = &model.User{
				ID:       0,
				Username: "未知用户",
			}
		}
		users = append(users, user)
	}

	return users, scores, nil
}

// ChangeScore 修改用户分数
func (repo *userRepository) ChangeScore(ctx context.Context, uid int64, delta int) {
	err := repo.cache.ChangeScore(ctx, uid, delta)
	if err != nil {
		slog.Error("Change User Score Failed", "error", err)
		return
	}
}
