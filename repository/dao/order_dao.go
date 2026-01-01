package dao

import (
	"context"
	"errors"

	"github.com/go-sql-driver/mysql"
	"github.com/yzletter/go-postery/model"
	"gorm.io/gorm"
)

type gormOrderDAO struct {
	db *gorm.DB
}

func NewOrderDAO(db *gorm.DB) OrderDAO {
	return &gormOrderDAO{db: db}
}

func (dao *gormOrderDAO) Create(ctx context.Context, order *model.Order) error {
	result := dao.db.WithContext(ctx).Model(&model.Order{}).Create(order)
	if result.Error != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(result.Error, &mysqlErr) && mysqlErr.Number == 1062 {
			return ErrUniqueKey
		}
		// 系统层面错误
		return ErrServerInternal
	}
	return nil
}
