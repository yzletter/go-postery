package dao

import (
	"context"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/yzletter/go-postery/backend/micro/user/model"
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

// GetProfile 根据 ID 查找用户资料
func (dao *gormUserDAO) GetProfile(ctx context.Context, id int64) (*model.Profile, error) {
	// 1. 构造结构体对象
	profile := &model.Profile{}

	// 2. 操作数据库
	result := dao.db.WithContext(ctx).
		Where("user_id = ? AND deleted_at IS NULL", id).
		First(profile)

	if result.Error != nil {
		// 业务层面错误
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, ErrServerInternal
	}

	// 3. 返回结果
	return profile, nil
}

// GetIDAfterTime 根据时间查找之后创建的用户 ID
func (dao *gormUserDAO) GetIDAfterTime(ctx context.Context, timeAt time.Time) ([]int64, error) {
	var ids []int64
	result := dao.db.WithContext(ctx).Model(&model.Profile{}).
		Where("created_at >= ? AND deleted_at IS NULL", timeAt).
		Pluck("user_id", &ids)

	if result.Error != nil {
		return nil, ErrServerInternal
	}
	return ids, nil
}

// UpdateProfile 根据 ID 修改用户资料的多个字段
func (dao *gormUserDAO) UpdateProfile(ctx context.Context, id int64, updates map[string]any) error {
	// 1. 操作数据库
	result := dao.db.WithContext(ctx).Model(&model.Profile{}).
		Where("user_id = ? AND deleted_at IS NULL", id).
		Updates(updates)

	if result.Error != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(result.Error, &mysqlErr) && mysqlErr.Number == 1062 {
			return ErrUniqueKey
		}
		return ErrServerInternal
	} else if result.RowsAffected == 0 {
		// 没有字段变化时 Gorm 也可能返回 0 行，回查确认用户资料是否存在
		var count int64
		err := dao.db.WithContext(ctx).Model(&model.Profile{}).
			Where("user_id = ? AND deleted_at IS NULL", id).
			Count(&count).Error
		if err != nil {
			return ErrServerInternal
		}
		if count == 0 {
			return ErrRecordNotFound
		}
	}
	// 2. 返回结果
	return nil
}
