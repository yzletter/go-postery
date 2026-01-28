package repository

import (
	"context"
	"log/slog"

	"github.com/yzletter/go-postery/user/model"
	"github.com/yzletter/go-postery/user/repository/cache"
	"github.com/yzletter/go-postery/user/repository/dao"
)

type userRepository struct {
	dao   dao.UserDAO
	cache cache.UserCache
}

func NewUserRepository(userDAO dao.UserDAO, userCache cache.UserCache) UserRepository {
	return &userRepository{dao: userDAO, cache: userCache}
}

// GetProfileByID 根据 ID 查找用户资料
func (repo *userRepository) GetProfileByID(ctx context.Context, uid int64) (*model.UserProfile, error) {
	userProfile, err := repo.dao.GetProfileByID(ctx, uid)
	if err != nil {
		return nil, toRepositoryErr(err)
	}

	return userProfile, nil
}

// UpdateProfile 根据 ID 修改用户资料的多个字段
func (repo *userRepository) UpdateProfile(ctx context.Context, id int64, updates map[string]any) error {
	err := repo.dao.UpdateProfile(ctx, id, updates)
	if err != nil {
		return toRepositoryErr(err)
	}

	// todo 更新 Cache

	return nil
}

// Top 返回热门推荐用户
func (repo *userRepository) Top(ctx context.Context) ([]*model.UserProfile, []float64, error) {
	ids, scores, err := repo.cache.Top(ctx)
	if err != nil {
		return nil, nil, toRepositoryErr(err)
	}

	var userProfiles []*model.UserProfile
	for _, id := range ids {
		userProfile, err := repo.dao.GetProfileByID(ctx, id)
		if err != nil {
			userProfile = &model.UserProfile{
				UserID:   0,
				Nickname: "未知用户",
			}
		}
		userProfiles = append(userProfiles, userProfile)
	}

	return userProfiles, scores, nil
}

// ChangeScore 修改用户分数
func (repo *userRepository) ChangeScore(ctx context.Context, uid int64, delta int) error {
	err := repo.cache.ChangeScore(ctx, uid, delta)
	if err != nil {
		slog.Error("Change User Score Failed", "error", err)
		return toRepositoryErr(err)
	}

	return nil
}
