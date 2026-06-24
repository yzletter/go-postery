package dao

import (
	"context"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/yzletter/go-postery/backend/micro/code/domain"
	"github.com/yzletter/go-postery/backend/micro/code/model"
	"gorm.io/gorm"
)

type gormCodeDAO struct {
	db *gorm.DB
}

func NewCodeDAO(db *gorm.DB) CodeDAO {
	return &gormCodeDAO{
		db: db,
	}
}

func (dao *gormCodeDAO) Create(ctx context.Context, code *model.VerificationCode) error {
	result := dao.db.WithContext(ctx).Model(&model.VerificationCode{}).Create(code)
	if result.Error != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(result.Error, &mysqlErr) && mysqlErr.Number == 1062 {
			return ErrUniqueKey
		}
		return ErrServerInternal
	}
	return nil
}

func (dao *gormCodeDAO) MarkVerified(ctx context.Context, biz domain.BizType, identifier string, codeHash string) error {
	now := time.Now()
	result := dao.db.WithContext(ctx).Model(&model.VerificationCode{}).
		Where("biz = ? AND identifier = ? AND code_hash = ? AND status = ? AND expires_at >= ?", biz, identifier, codeHash, model.CodeStatusSent, now).
		Order("created_at DESC").Limit(1).
		Updates(map[string]interface{}{
			"status":      model.CodeStatusVerified,
			"verified_at": now,
		})
	if result.Error != nil {
		return ErrServerInternal
	}

	if result.RowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}
