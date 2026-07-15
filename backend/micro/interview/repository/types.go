package repository

import (
	"context"

	"github.com/yzletter/go-postery/backend/micro/interview/model"
)

type InterviewRepository interface {
	// SaveProfile 保存用户画像
	//
	// Parameter:
	//	- profile: 用户画像
	//
	// Return:
	//	- error: 可能返回的错误
	SaveProfile(ctx context.Context, profile *model.InterviewProfile) error

	// LoadProfile 加载用户画像
	//
	// Parameter:
	//	- userID: 用户 ID
	//
	// Return:
	//	- *model.InterviewProfile: 用户画像
	//	- error: 可能返回的错误
	LoadProfile(ctx context.Context, userID int64) (*model.InterviewProfile, error)

	UpsertSession(ctx context.Context, userID int64, sessionID int64, data []byte) error

	LoadSession(ctx context.Context, sessionID int64) ([]byte, error)
}
