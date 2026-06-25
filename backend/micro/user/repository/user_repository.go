package repository

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/yzletter/go-postery/backend/micro/user/domain"
	"github.com/yzletter/go-postery/backend/micro/user/repository/cache"
	"github.com/yzletter/go-postery/backend/micro/user/repository/dao"
)

type userRepository struct {
	dao   dao.UserDAO
	cache cache.UserCache
}

// NewUserRepository 构造函数
func NewUserRepository(userDAO dao.UserDAO, userCache cache.UserCache) UserRepository {
	return &userRepository{dao: userDAO, cache: userCache}
}

// userRepositoryLogger 构造用户仓储日志
func userRepositoryLogger(method string) *slog.Logger {
	return slog.With("component", "user_repository", "method", method)
}

// GetProfile 根据 ID 查找用户资料
func (repo *userRepository) GetProfile(ctx context.Context, id int64) (domain.Profile, error) {
	logger := userRepositoryLogger("GetProfile").With("user_id", id)

	// 先查缓存
	if cachedProfile, err := repo.cache.GetProfile(ctx, id); err == nil {
		logger.Debug("profile cache hit")
		return cachedProfile, nil
	} else if !errors.Is(err, redis.Nil) {
		logger.Warn("get profile cache failed", "error", err)
	}

	// 缓存未命中时查数据库
	profile, err := repo.dao.GetProfile(ctx, id)
	if err != nil {
		return domain.Profile{}, toRepositoryErr(err)
	}

	// model 转 domain
	back := domain.ToDomainProfile(profile)

	// 更新缓存
	if err := repo.cache.SetProfile(ctx, id, back); err != nil {
		logger.Warn("set profile cache failed", "error", err)
	}

	return back, nil
}

// UpdateProfile 根据 ID 修改用户资料的多个字段
func (repo *userRepository) UpdateProfile(ctx context.Context, id int64, updates map[string]any) error {
	logger := userRepositoryLogger("UpdateProfile").With("user_id", id)

	// 写数据库
	if err := repo.dao.UpdateProfile(ctx, id, updates); err != nil {
		return toRepositoryErr(err)
	}

	// 删缓存
	if err := repo.cache.DelProfile(ctx, id); err != nil {
		logger.Warn("delete profile cache failed", "error", err)
	}

	return nil
}

// GetIDAfterTime 根据时间查找之后创建的用户 ID
func (repo *userRepository) GetIDAfterTime(ctx context.Context, timeAfter time.Time) ([]int64, error) {
	ids, err := repo.dao.GetIDAfterTime(ctx, timeAfter)
	if err != nil {
		return nil, toRepositoryErr(err)
	}
	return ids, nil
}
