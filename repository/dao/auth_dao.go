package dao

import (
	"context"
	"errors"
	"log/slog"

	"github.com/go-sql-driver/mysql"
	"github.com/yzletter/go-postery/errno"
	"github.com/yzletter/go-postery/model"
	"gorm.io/gorm"
)

type gormAuthDAO struct {
	db *gorm.DB
}

func NewAuthDAO(db *gorm.DB) AuthDAO {
	return &gormAuthDAO{
		db: db,
	}
}

// CreateUser 创建用户（包括用户最小项、用户登录认证、用户密码、用户资料、注册扩展功能）
func (dao *gormAuthDAO) CreateUser(ctx context.Context, authAggregate *model.AuthAggregate) error {
	err := dao.db.WithContext(ctx).Transaction(
		func(tx *gorm.DB) error {
			if err := tx.Create(authAggregate.User).Error; err != nil {
				return err
			}
			if err := tx.Create(authAggregate.UserProfile).Error; err != nil {
				return err
			}
			if err := tx.Create(authAggregate.AuthIdentity).Error; err != nil {
				return err
			}

			// 写 OutBox
			for _, event := range authAggregate.Events {
				if err := tx.Create(event).Error; err != nil {
					return err
				}
			}

			if authAggregate.AuthPassword == nil {
				return nil
			}
			if err := tx.Create(authAggregate.AuthPassword).Error; err != nil {
				return err
			}
			return nil
		})
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return ErrUniqueKey
		}
		return ErrServerInternal
	}
	return nil
}

// GetAuthIdentity 根据登录方式和凭证获取登录认证
func (dao *gormAuthDAO) GetAuthIdentity(ctx context.Context, authType int, identifier string) (*model.AuthIdentity, error) {
	var authIdentity model.AuthIdentity
	result := dao.db.WithContext(ctx).Model(&model.AuthIdentity{}).Where("auth_type = ? AND identifier = ?", authType, identifier).First(&authIdentity)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}

		slog.Error(FindFailed, "auth_type", authType, "identifier", identifier, "error", result.Error)
		return nil, ErrServerInternal
	}

	return &authIdentity, nil
}

// GetAuthIdentityByIdentifier 根据凭证获取登录认证
func (dao *gormAuthDAO) GetAuthIdentityByIdentifier(ctx context.Context, identifier string) (*model.AuthIdentity, error) {
	var authIdentity model.AuthIdentity
	result := dao.db.WithContext(ctx).Model(&model.AuthIdentity{}).Where("identifier = ? AND is_verified = ?", identifier, 1).First(&authIdentity)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}

		slog.Error(FindFailed, "identifier", identifier, "error", result.Error)
		return nil, ErrServerInternal
	}

	return &authIdentity, nil
}

// GetAuthIdentityByAuthType 根据认证方式获取登录认证
func (dao *gormAuthDAO) GetAuthIdentityByAuthType(ctx context.Context, uid int64, authType int) (*model.AuthIdentity, error) {
	var authIdentity model.AuthIdentity
	result := dao.db.WithContext(ctx).Model(&model.AuthIdentity{}).Where("user_id = ? AND auth_type = ? AND is_verified = ?", uid, authType, 1).First(&authIdentity)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}

		slog.Error(FindFailed, "uid", uid, "auth_type", authType, "error", result.Error)
		return nil, ErrServerInternal
	}

	return &authIdentity, nil
}

// GetAuthIdentityByUID 获取用户身份认证
func (dao *gormAuthDAO) GetAuthIdentityByUID(ctx context.Context, uid int64) ([]*model.AuthIdentity, error) {
	var authIdentity []*model.AuthIdentity
	result := dao.db.WithContext(ctx).Model(&model.AuthIdentity{}).Where("user_id = ? AND is_verified = ?", uid, 1).Find(&authIdentity)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}

		slog.Error(FindFailed, "uid", uid, "error", result.Error)
		return nil, ErrServerInternal
	}

	return authIdentity, nil
}

// GetPasswordHash 根据 UID 获取用户密码
func (dao *gormAuthDAO) GetPasswordHash(ctx context.Context, uid int64) (string, error) {
	var authPassword model.AuthPassword
	result := dao.db.WithContext(ctx).Model(&model.AuthPassword{}).Where("user_id = ?", uid).First(&authPassword)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return "", ErrRecordNotFound
		}

		slog.Error(FindFailed, "uid", uid, "error", result.Error)
		return "", ErrServerInternal
	}

	return authPassword.PasswordHash, nil
}

// HasPassword 查询密码状态
func (dao *gormAuthDAO) HasPassword(ctx context.Context, uid int64) (bool, error) {
	var cnt int64
	result := dao.db.Model(&model.AuthPassword{}).WithContext(ctx).Where("user_id = ?", uid).Count(&cnt)
	if result.Error != nil {
		slog.Error(FindFailed, "uid", uid, "error", result.Error)
		return false, errno.ErrServerInternal
	}

	return cnt > 0, nil
}

// SetPassword 初始化密码
func (dao *gormAuthDAO) SetPassword(ctx context.Context, authPassword *model.AuthPassword) error {
	result := dao.db.WithContext(ctx).Create(authPassword)
	if result.Error != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(result.Error, &mysqlErr) && mysqlErr.Number == 1062 {
			return ErrUniqueKey
		}
		return ErrServerInternal
	}

	return nil
}

// UpdatePasswordHash 修改用户密码
func (dao *gormAuthDAO) UpdatePasswordHash(ctx context.Context, uid int64, passwordHash string) error {
	// 先查是否有密码
	var cnt int64
	if err := dao.db.WithContext(ctx).Model(&model.AuthPassword{}).Where("user_id = ?", uid).Count(&cnt).Error; err != nil {
		return ErrServerInternal
	}
	if cnt == 0 {
		return ErrRecordNotFound
	}

	result := dao.db.WithContext(ctx).Model(&model.AuthPassword{}).Where("user_id = ?", uid).Update("password_hash", passwordHash)
	if result.Error != nil {
		slog.Error(UpdateFailed, "uid", uid, "error", result.Error)
		return ErrServerInternal
	}

	return nil
}
