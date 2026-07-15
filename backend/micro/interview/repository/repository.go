package repository

import (
	"context"
	"log/slog"

	"github.com/yzletter/go-postery/backend/micro/interview/model"
	"github.com/yzletter/go-postery/backend/micro/interview/repository/cache"
	"github.com/yzletter/go-postery/backend/micro/interview/repository/dao"
)

type interviewRepository struct {
	dao   dao.InterviewDAO
	cache cache.InterviewCache
}

// NewInterviewRepository 构造函数
func NewInterviewRepository(interviewDAO dao.InterviewDAO, interviewCache cache.InterviewCache) InterviewRepository {
	return &interviewRepository{
		dao:   interviewDAO,
		cache: interviewCache,
	}
}

// SaveProfile 保存用户画像
func (repo *interviewRepository) SaveProfile(ctx context.Context, profile *model.InterviewProfile) error {
	if profile == nil {
		return ErrParamsInvalid
	}

	// 写 MySQL
	if err := repo.dao.SaveProfile(ctx, profile); err != nil {
		return toRepositoryErr(err)
	}

	// 删除缓存，后续读取时回源重建
	if err := repo.cache.DelProfile(ctx, profile.UserID); err != nil {
		slog.Warn("delete interview profile cache failed", "user_id", profile.UserID, "error", err)
	}

	return nil
}

// LoadProfile 加载用户画像
func (repo *interviewRepository) LoadProfile(ctx context.Context, userID int64) (*model.InterviewProfile, error) {
	// 查缓存
	if profile, err := repo.cache.GetProfile(ctx, userID); err == nil {
		return profile, nil
	}

	// 查 MySQL
	profile, err := repo.dao.GetProfile(ctx, userID)
	if err != nil {
		return nil, toRepositoryErr(err)
	}

	// 写缓存
	if err := repo.cache.SetProfile(ctx, profile); err != nil {
		slog.Warn("set interview profile cache failed", "user_id", userID, "error", err)
	}

	return profile, nil
}

// UpsertSession 根据会话 ID 保存或更新面试会话快照
func (repo *interviewRepository) UpsertSession(ctx context.Context, userID int64, sessionID int64, data []byte) error {
	// 写 MySQL
	if err := repo.dao.UpsertSession(ctx, userID, sessionID, data); err != nil {
		return toRepositoryErr(err)
	}

	// 写缓存，过期时间由 cache 层统一维护
	if err := repo.cache.SetSession(ctx, sessionID, data); err != nil {
		slog.Warn("set interview session cache failed", "session_id", sessionID, "error", err)
	}
	return nil
}

// LoadSession 根据会话 ID 加载面试会话快照
func (repo *interviewRepository) LoadSession(ctx context.Context, sessionID int64) ([]byte, error) {
	// 查缓存
	if data, err := repo.cache.GetSession(ctx, sessionID); err == nil {
		return data, nil
	}

	// 查 MySQL
	data, err := repo.dao.GetSession(ctx, sessionID)
	if err != nil {
		return nil, toRepositoryErr(err)
	}

	// 写缓存
	if err := repo.cache.SetSession(ctx, sessionID, data); err != nil {
		slog.Warn("set interview session cache failed", "session_id", sessionID, "error", err)
	}
	return data, nil
}
