package dao

import (
	"context"
	"errors"

	"github.com/go-sql-driver/mysql"
	"github.com/yzletter/go-postery/backend/micro/interview/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// gormInterviewDAO 用 Gorm 实现 InterviewDAO
type gormInterviewDAO struct {
	db *gorm.DB
}

// NewInterviewDAO 构造函数
func NewInterviewDAO(db *gorm.DB) InterviewDAO {
	return &gormInterviewDAO{db: db}
}

// SaveProfile 保存用户画像
func (dao *gormInterviewDAO) SaveProfile(ctx context.Context, profile *model.InterviewProfile) error {
	// 0. 兜底
	if profile == nil || profile.UserID == 0 {
		return ErrParamsInvalid
	}

	// 1. 操作数据库
	result := dao.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"name",
			"skill_level",
			"weak_points",
			"updated_at",
		}),
	}).Create(profile)
	if result.Error != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(result.Error, &mysqlErr) && mysqlErr.Number == 1062 {
			return ErrUniqueKey
		}
		return ErrServerInternal
	}

	// 2. 返回结果
	return nil
}

// GetProfile 根据用户 ID 查询用户画像
func (dao *gormInterviewDAO) GetProfile(ctx context.Context, userID int64) (*model.InterviewProfile, error) {
	// 0. 兜底
	if userID == 0 {
		return nil, ErrParamsInvalid
	}

	// 1. 操作数据库
	var profile model.InterviewProfile
	result := dao.db.WithContext(ctx).Model(&model.InterviewProfile{}).Where("id = ?", userID).First(&profile)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, ErrServerInternal
	}

	// 2. 返回结果
	return &profile, nil
}

// UpsertSession 保存面试会话快照
func (dao *gormInterviewDAO) UpsertSession(ctx context.Context, userID int64, sessionID int64, data []byte) error {
	// 0. 兜底
	if sessionID == 0 {
		return ErrParamsInvalid
	}

	// 1. 操作数据库
	result := dao.db.WithContext(ctx).Model(&model.InterviewSession{}).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.Assignments(
				map[string]interface{}{
					"session_data": data,
				}),
		}).
		Create(&model.InterviewSession{
			ID:          sessionID,
			UserID:      userID,
			SessionData: data,
		})
	if result.Error != nil {
		return ErrServerInternal
	}

	if result.RowsAffected == 0 {
		// 数据未变化时 MySQL 也可能返回 0 行，回查确认会话是否存在
		var cnt int64
		err := dao.db.WithContext(ctx).Model(&model.InterviewSession{}).Where("id = ?", sessionID).Count(&cnt).Error
		if err != nil {
			return ErrServerInternal
		}
		if cnt == 0 {
			return ErrRecordNotFound
		}
	}

	// 2. 返回结果
	return nil
}

// GetSession 根据会话 ID 查询面试会话快照
func (dao *gormInterviewDAO) GetSession(ctx context.Context, sessionID int64) ([]byte, error) {
	// 0. 兜底
	if sessionID == 0 {
		return nil, ErrParamsInvalid
	}

	// 1. 操作数据库
	var session model.InterviewSession
	result := dao.db.WithContext(ctx).Model(&model.InterviewSession{}).
		Where("id = ?", sessionID).
		First(&session)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, ErrServerInternal
	}

	// 2. 返回结果
	return session.SessionData, nil
}
