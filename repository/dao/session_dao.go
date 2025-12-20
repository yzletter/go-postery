package dao

import (
	"context"
	"errors"
	"log/slog"

	"github.com/go-sql-driver/mysql"
	"github.com/yzletter/go-postery/model"
	"gorm.io/gorm"
)

type gormSessionDAO struct {
	db *gorm.DB
}

func NewSessionDAO(db *gorm.DB) SessionDAO {
	return &gormSessionDAO{db: db}
}

func (dao *gormSessionDAO) Create(ctx context.Context, session *model.Session) error {
	result := dao.db.WithContext(ctx).Create(session)
	if result.Error != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(result.Error, &mysqlErr) {
			return ErrUniqueKey
		}
		return ErrServerInternal
	}

	return nil
}

func (dao *gormSessionDAO) GetByUidAndTargetID(ctx context.Context, uid, targetID int64) (*model.Session, error) {
	var session *model.Session
	result := dao.db.WithContext(ctx).Model(&model.Session{}).Where("user_id = ? AND target_id = ? AND deleted_at IS NULL", uid, targetID).First(&session)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		slog.Error(FindFailed, "user_id", uid, "error", result.Error)
		return nil, ErrServerInternal
	}

	return session, nil
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
