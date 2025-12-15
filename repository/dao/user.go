package dao

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/yzletter/go-postery/infra/snowflake"
	"github.com/yzletter/go-postery/model"
	"gorm.io/gorm"
)

// GormUserDAO 用 Gorm 实现 UserDAO
type GormUserDAO struct {
	db *gorm.DB
}

// NewUserDAO 构造函数
func NewUserDAO(db *gorm.DB) *GormUserDAO {
	return &GormUserDAO{
		db: db,
	}
}

// Create 创建 User
func (dao *GormUserDAO) Create(ctx context.Context, user *model.User) (*model.User, error) {
	// 0. 技术字段完整性保证
	if user.ID == 0 {
		user.ID = snowflake.NextID()
	}
	if user.Username == "" || user.Email == "" || user.PasswordHash == "" {
		return nil, ErrParamsInvalid
	}

	// 1. 操作数据库
	result := dao.db.WithContext(ctx).Create(user)
	if result.Error != nil {
		// 业务层面错误
		var mysqlErr *mysql.MySQLError
		if errors.As(result.Error, &mysqlErr) && mysqlErr.Number == 1062 { // 判断是否为 Unique Key 冲突
			return nil, ErrUniqueKeyConflict
		}

		// 系统层面错误
		slog.Error(CreateFailed, "username", user.Username, "error", result.Error)
		return nil, ErrInternal
	}

	// 2. 返回结果
	return user, nil
}

// Delete 软删除 User
func (dao *GormUserDAO) Delete(ctx context.Context, id int64) error {
	// 1. 操作数据库
	now := time.Now()
	result := dao.db.WithContext(ctx).Model(&model.User{}).Where("id = ? AND deleted_at IS NULL", id).Update("deleted_at", &now)
	if result.Error != nil {
		// 系统层面错误
		slog.Error(DeleteFailed, "id", id, "error", result.Error)
		return ErrInternal
	} else if result.RowsAffected == 0 {
		// 业务层面错误
		return ErrRecordNotFound
	}

	// 2. 返回结果
	return nil
}

// GetPasswordHash 返回 User 的 PasswordHash
func (dao *GormUserDAO) GetPasswordHash(ctx context.Context, id int64) (string, error) {
	// 1. 操作数据库
	var res string
	result := dao.db.WithContext(ctx).Model(&model.User{}).Select("password_hash").Where("id = ? AND deleted_at IS NULL", id).Take(&res)
	if result.Error != nil {
		// 业务层面错误
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return "", ErrRecordNotFound
		}
		// 系统层面错误
		slog.Error(FindFailed, "id", id, "error", result.Error)
		return "", ErrInternal
	}

	// 2. 返回结果
	return res, nil
}

// GetStatus 返回 User 的 Status
func (dao *GormUserDAO) GetStatus(ctx context.Context, id int64) (int, error) {
	// 1. 操作数据库
	var status int
	result := dao.db.WithContext(ctx).Model(&model.User{}).Select("status").Where("id = ? AND deleted_at IS NULL", id).Take(&status)
	if result.Error != nil {
		// 业务层面错误
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return 0, ErrRecordNotFound
		}
		// 系统层面错误
		slog.Error(FindFailed, "id", id, "error", result.Error)
		return 0, ErrInternal
	}

	// 2. 返回结果
	return status, nil
}

// GetByID 根据 User 的 ID 查找不带密码的 User
func (dao *GormUserDAO) GetByID(ctx context.Context, id int64) (*model.User, error) {
	// 1. 构造结构体对象
	user := &model.User{}

	// 2. 操作数据库
	result := dao.db.WithContext(ctx).Omit("password_hash").Where("id = ? AND deleted_at IS NULL", id).First(user)
	if result.Error != nil {
		// 业务层面错误
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		// 系统层面错误
		slog.Error(FindFailed, "id", id, "error", result.Error)
		return nil, ErrInternal
	}

	// 3. 返回结果
	return user, nil
}

// GetByUsername 根据 User 的 Username 查找带密码的 User
func (dao *GormUserDAO) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	// 1. 构造结构体对象
	user := &model.User{}

	// 2. 操作数据库
	result := dao.db.WithContext(ctx).Where("username = ? AND deleted_at IS NULL", username).First(user)
	if result.Error != nil {
		// 业务层面错误
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		// 系统层面错误
		slog.Error(FindFailed, "username", username, "error", result.Error)
		return nil, ErrInternal
	}

	// 3. 返回结果
	return user, nil
}

// UpdatePasswordHash 更新 User 的 PasswordHash
func (dao *GormUserDAO) UpdatePasswordHash(ctx context.Context, id int64, newHash string) error {
	// 1. 操作数据库
	result := dao.db.WithContext(ctx).WithContext(ctx).Model(&model.User{}).Where("id = ? AND deleted_at IS NULL", id).Update("password_hash", newHash)
	if result.Error != nil {
		// 系统层面错误
		slog.Error(UpdateFailed, "id", id, "error", result.Error)
		return ErrInternal
	} else if result.RowsAffected == 0 {
		// 业务层面错误
		var cnt int64
		result2 := dao.db.WithContext(ctx).Model(&model.User{}).Where("id = ? AND deleted_at IS NULL", id).Count(&cnt)
		if result2.Error != nil {
			// 系统层面错误
			slog.Error(FindFailed, "id", id, "error", result.Error)
			return ErrInternal
		}

		if cnt == 0 {
			// 记录不存在
			return ErrRecordNotFound
		}
	}

	// 2. 返回结果
	return nil
}

// UpdateProfile 更新 User 的多个字段
func (dao *GormUserDAO) UpdateProfile(ctx context.Context, id int64, updates map[string]any) error {
	// 1. 操作数据库
	result := dao.db.WithContext(ctx).Model(&model.User{}).Where("id = ? AND deleted_at IS NULL", id).Updates(updates)
	if result.Error != nil {
		// 系统层面错误
		slog.Error(UpdateFailed, "id", id, "error", result.Error)
		return ErrInternal
	} else if result.RowsAffected == 0 {
		// 业务层面错误
		var cnt int64
		result2 := dao.db.WithContext(ctx).Model(&model.User{}).Where("id = ? AND deleted_at IS NULL", id).Count(&cnt)
		if result2.Error != nil {
			// 系统层面错误
			slog.Error(FindFailed, "id", id, "error", result.Error)
			return ErrInternal
		}

		if cnt == 0 {
			// 记录不存在
			return ErrRecordNotFound
		}
	}

	// 2. 返回结果
	return nil
}
