package dao

import (
	"context"
	"log/slog"

	"github.com/yzletter/go-postery/model"
	"gorm.io/gorm"
)

type gormSessionDAO struct {
	db *gorm.DB
}

func NewSessionDAO(db *gorm.DB) SessionDAO {
	return &gormSessionDAO{db: db}
}

func (dao *gormSessionDAO) GetByUid(ctx context.Context, uid int64) ([]*model.Session, error) {
	var sessions []*model.Session
	result := dao.db.WithContext(ctx).Model(&model.Session{}).Where("user_id = ? AND deleted_at IS NULL", uid).Order("updated_at DESC").Find(&sessions)
	if result.Error != nil {
		slog.Error(FindFailed, "user_id", uid, "error", result.Error)
		return nil, ErrServerInternal
	}

	return sessions, nil
}
