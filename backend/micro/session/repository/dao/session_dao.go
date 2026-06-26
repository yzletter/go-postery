package dao

import (
	"context"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/yzletter/go-postery/backend/micro/session/model"
	"gorm.io/gorm"
)

type gormSessionDAO struct {
	db *gorm.DB
}

func NewSessionDAO(db *gorm.DB) SessionDAO {
	return &gormSessionDAO{db: db}
}

func (dao *gormSessionDAO) Create(ctx context.Context, sessions ...*model.Session) error {
	err := dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, session := range sessions {
			result := tx.Create(session)
			if result.Error != nil {
				var mysqlErr *mysql.MySQLError
				if errors.As(result.Error, &mysqlErr) && mysqlErr.Number == 1062 {
					return ErrUniqueKey
				}
				return ErrServerInternal
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (dao *gormSessionDAO) Recover(ctx context.Context, uid, targetID int64) (*model.Session, error) {
	var session model.Session
	err := dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 恢复当前用户侧软删除的会话
		result := tx.Model(&model.Session{}).
			Where("user_id = ? AND target_id = ? AND deleted_at IS NOT NULL", uid, targetID).
			Update("deleted_at", nil)
		if result.Error != nil {
			return ErrServerInternal
		}
		if result.RowsAffected == 0 {
			return ErrRecordNotFound
		}

		// 查出恢复后的会话
		result = tx.Model(&model.Session{}).
			Where("user_id = ? AND target_id = ? AND deleted_at IS NULL", uid, targetID).
			First(&session)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return ErrRecordNotFound
			}
			return ErrServerInternal
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (dao *gormSessionDAO) GetByUidAndTargetID(ctx context.Context, uid, targetID int64) (*model.Session, error) {
	var session *model.Session
	result := dao.db.WithContext(ctx).Model(&model.Session{}).Where("user_id = ? AND target_id = ? AND deleted_at IS NULL", uid, targetID).First(&session)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, ErrServerInternal
	}

	return session, nil
}

func (dao *gormSessionDAO) GetByUid(ctx context.Context, uid int64) ([]*model.Session, error) {
	var sessions []*model.Session
	result := dao.db.WithContext(ctx).Model(&model.Session{}).Where("user_id = ? AND deleted_at IS NULL", uid).Order("updated_at DESC").Find(&sessions)
	if result.Error != nil {
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
		return nil, ErrServerInternal
	}

	return session, nil
}

func (dao *gormSessionDAO) Delete(ctx context.Context, uid, sid int64) error {
	now := time.Now()
	result := dao.db.WithContext(ctx).Model(&model.Session{}).Where("user_id = ? AND session_id = ? AND deleted_at IS NULL", uid, sid).Update("deleted_at", &now)
	if result.Error != nil {
		return ErrServerInternal
	} else if result.RowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

func (dao *gormSessionDAO) UpdateUnread(ctx context.Context, uid int64, sid int64, updates model.UpdateUnread) error {
	result := dao.db.WithContext(ctx).Model(&model.Session{}).Where("user_id = ? AND session_id = ? AND deleted_at IS NULL", uid, sid).
		Updates(updates.Updates).UpdateColumn("unread_count", gorm.Expr("unread_count + ?", updates.Delta))
	if result.Error != nil {
		return ErrServerInternal
	}

	if result.RowsAffected == 0 {
		// 会话已删除，进行恢复
		result2 := dao.db.WithContext(ctx).Model(&model.Session{}).Where("user_id = ? AND session_id = ? AND deleted_at IS NOT NULL", uid, sid).
			Update("deleted_at", nil).Updates(updates.Updates).UpdateColumn("unread_count", gorm.Expr("unread_count + ?", updates.Delta))
		if result2.Error != nil {
			return ErrServerInternal
		}
		if result2.RowsAffected == 0 {
			return ErrRecordNotFound
		}
	}

	return nil
}

func (dao *gormSessionDAO) ClearUnread(ctx context.Context, uid int64, sid int64) error {
	result := dao.db.WithContext(ctx).Model(&model.Session{}).Where("user_id = ? AND session_id = ? AND deleted_at IS NULL", uid, sid).
		Update("unread_count", 0)
	if result.Error != nil {
		return ErrServerInternal
	} else if result.RowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}
