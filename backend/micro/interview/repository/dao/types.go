package dao

import (
	"context"

	"github.com/yzletter/go-postery/backend/micro/interview/model"
)

type InterviewDAO interface {
	SaveProfile(ctx context.Context, profile *model.InterviewProfile) error

	GetProfile(ctx context.Context, userID int64) (*model.InterviewProfile, error)

	UpsertSession(ctx context.Context, userID int64, sessionID int64, data []byte) error

	GetSession(ctx context.Context, sessionID int64) ([]byte, error)
}
