package cache

import (
	"context"

	"github.com/yzletter/go-postery/backend/micro/interview/model"
)

type InterviewCache interface {
	// SetProfile 设置用户画像缓存
	SetProfile(ctx context.Context, profile *model.InterviewProfile) error

	// GetProfile 获取用户画像缓存
	GetProfile(ctx context.Context, userID int64) (*model.InterviewProfile, error)

	// DelProfile 删除用户画像缓存
	DelProfile(ctx context.Context, userID int64) error

	// SetSession 设置面试会话缓存
	SetSession(ctx context.Context, sessionID int64, data []byte) error

	// GetSession 获取面试会话缓存
	GetSession(ctx context.Context, sessionID int64) ([]byte, error)

	// DelSession 删除面试会话缓存
	DelSession(ctx context.Context, sessionID int64) error
}
