package dao

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/yzletter/go-postery/dto/session"
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

func (dao *gormSessionDAO) GetByID(ctx context.Context, uid, sid int64) (*model.Session, error) {
	var session *model.Session
	result := dao.db.WithContext(ctx).Model(&model.Session{}).Where("user_id = ? AND session_id = ? AND deleted_at IS NULL", uid, sid).First(&session)
	if result.Error != nil {
		// 业务层面错误
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		// 系统层面错误
		slog.Error(FindFailed, "user_id", uid, "session_id", sid, "error", result.Error)
		return nil, ErrServerInternal
	}

	return session, nil
}

func (dao *gormSessionDAO) Delete(ctx context.Context, uid, sid int64) error {
	now := time.Now()
	result := dao.db.WithContext(ctx).Model(&model.Session{}).Where("user_id = ? AND session_id = ? AND deleted_at IS NULL", uid, sid).Update("deleted_at", &now)
	if result.Error != nil {
		// 系统层面错误
		slog.Error(DeleteFailed, "user_id", uid, "session_id", sid, "error", result.Error)
		return ErrServerInternal
	} else if result.RowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

func (dao *gormSessionDAO) UpdateUnread(ctx context.Context, uid int64, sid int64, updates session.UpdateUnreadRequest) error {
	result := dao.db.WithContext(ctx).Model(&model.Session{}).Where("user_id = ? AND session_id = ? AND deleted_at IS NULL", uid, sid).
		Updates(updates.Updates).UpdateColumn("unread_count", gorm.Expr("unread_count + ?", updates.Delta))
	if result.Error != nil {
		slog.Error(UpdateFailed, "error", result.Error)
		return ErrServerInternal
	}

	if result.RowsAffected == 0 {
		// 会话已删除，进行恢复 todo 考虑
		result2 := dao.db.WithContext(ctx).Model(&model.Session{}).Where("user_id = ? AND session_id = ? AND deleted_at IS NOT NULL", uid, sid).
			Update("deleted_at", nil).Updates(updates.Updates).UpdateColumn("unread_count", gorm.Expr("unread_count + ?", updates.Delta))
		if result2.Error != nil {
			slog.Error(UpdateFailed, "error", result2.Error)
			return ErrServerInternal
		}
	}

	return nil
}

func (dao *gormSessionDAO) ClearUnread(ctx context.Context, uid int64, sid int64) error {
	result := dao.db.WithContext(ctx).Model(&model.Session{}).Where("user_id = ? AND session_id = ? AND deleted_at IS NULL", uid, sid).
		Update("unread_count", 0)
	if result.Error != nil {
		slog.Error(UpdateFailed, "error", result.Error)
		return ErrServerInternal
	} else if result.RowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}
