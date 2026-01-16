package dao

import (
	"context"
	"errors"
	"log/slog"

	"github.com/yzletter/go-postery/model"
	"gorm.io/gorm"
)

// gormUserDAO 用 Gorm 实现 UserDAO
type gormUserDAO struct {
	db *gorm.DB
}

// NewUserDAO 构造函数
func NewUserDAO(db *gorm.DB) UserDAO {
	return &gormUserDAO{
		db: db,
	}
}

// GetProfileByID 根据 ID 查找用户资料
func (dao *gormUserDAO) GetProfileByID(ctx context.Context, id int64) (*model.UserProfile, error) {
	// 1. 构造结构体对象
	userProfile := &model.UserProfile{}

	// 2. 操作数据库
	result := dao.db.WithContext(ctx).Where("user_id = ? AND deleted_at IS NULL", id).First(userProfile)
	if result.Error != nil {
		// 业务层面错误
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		// 系统层面错误
		slog.Error(FindFailed, "id", id, "error", result.Error)
		return nil, ErrServerInternal
	}

	// 3. 返回结果
	return userProfile, nil
}

// UpdateProfile 根据 ID 修改用户资料的多个字段
func (dao *gormUserDAO) UpdateProfile(ctx context.Context, id int64, updates map[string]any) error {
	// 1. 操作数据库
	result := dao.db.WithContext(ctx).Model(&model.UserProfile{}).Where("user_id = ? AND deleted_at IS NULL", id).Updates(updates)
	if result.Error != nil {
		// 系统层面错误
		slog.Error(UpdateFailed, "id", id, "error", result.Error)
		return ErrServerInternal
	} else if result.RowsAffected == 0 {
		// 业务层面错误
		var cnt int64
		result2 := dao.db.WithContext(ctx).Model(&model.UserProfile{}).Where("id = ? AND deleted_at IS NULL", id).Count(&cnt)
		if result2.Error != nil {
			// 系统层面错误
			slog.Error(FindFailed, "id", id, "error", result.Error)
			return ErrServerInternal
		}

		if cnt == 0 {
			// 记录不存在
			return ErrRecordNotFound
		}
	}

	// 2. 返回结果
	return nil
}
