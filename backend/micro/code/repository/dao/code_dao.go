package dao

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/go-sql-driver/mysql"
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
		slog.Error(CreateFailed, "biz", code.Biz, "identifier", code.Identifier, "error", result.Error)
		return ErrServerInternal
	}
	return nil
}

func (dao *gormCodeDAO) MarkVerified(ctx context.Context, biz int, identifier string, codeHash string) error {
	now := time.Now()
	result := dao.db.WithContext(ctx).Model(&model.VerificationCode{}).
		Where("biz = ? AND identifier = ? AND code_hash = ? AND status = ? AND expires_at >= ?", biz, identifier, codeHash, model.CodeStatusSent, now).
		Order("created_at DESC").Limit(1).
		Updates(map[string]interface{}{
			"status":      model.CodeStatusVerified,
			"verified_at": now,
		})
	if result.Error != nil {
		slog.Error(UpdateFailed, "biz", biz, "identifier", identifier, "error", result.Error)
		return ErrServerInternal
	}
	if result.RowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}
