package dao

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/yzletter/go-postery/microservice-backend/lottery/model"
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

func (dao *gormOrderDAO) Get(ctx context.Context, uid int64) (*model.Order, error) {
	var order *model.Order
	result := dao.db.WithContext(ctx).Model(&model.Order{}).Where("user_id = ? AND deleted_at IS NULL", uid).Order("created_at DESC").First(&order)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, ErrServerInternal
	}
	return order, nil
}

func (dao *gormOrderDAO) GetTempOrder(ctx context.Context, uid int64) (*model.Order, error) {
	var order *model.Order
	result := dao.db.WithContext(ctx).Model(&model.Order{}).Where("user_id = ? AND status = ? AND expire_at >= ? AND deleted_at IS NULL", uid, model.OrderStatusPending, time.Now()).Order("created_at DESC").First(&order)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, ErrServerInternal
	}
	return order, nil
}

// CreateTempOrder 创建临时订单
func (dao *gormOrderDAO) CreateTempOrder(ctx context.Context, order *model.Order) error {
	// 兜底
	if order.Status != model.OrderStatusPending {
		slog.Error("order status is invalid")
		order.Status = model.OrderStatusPending
	}

	if result := dao.db.WithContext(ctx).Model(&model.Order{}).Create(order); result.Error != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(result.Error, &mysqlErr) && mysqlErr.Number == 1062 {
			return ErrUniqueKey
		}
		// 系统层面错误
		return ErrServerInternal
	}
	return nil
}
func (dao *gormOrderDAO) RecycleTempOrder(ctx context.Context, uid int64, orderID int64) (bool, error) {
	// 1. Pending -> Expired
	now := time.Now()
	result := dao.db.WithContext(ctx).Model(&model.Order{}).
		Where("id = ? AND status = ? AND expire_at <= ? AND stock_rollback_status = ? AND deleted_at IS NULL",
			orderID, model.OrderStatusPending, now, model.StockRollbackStatusPending).
		Updates(map[string]interface{}{
			"status": model.OrderStatusExpired,
		})

	if result.Error != nil {
		return false, ErrServerInternal
	} else if result.RowsAffected == 0 {
		return false, ErrRecordNotFound
	}
	return true, nil
}

func (dao *gormOrderDAO) PayTempOrder(ctx context.Context, orderID int64) error {
	err := dao.db.WithContext(ctx).Transaction(
		func(tx *gorm.DB) error {
			// 1. 查询订单
			var order model.Order
			if err := tx.Where("id = ? AND deleted_at IS NULL", orderID).First(&order).Error; err != nil {
				return ErrRecordNotFound
			}

			// 支付回调幂等：已经支付过，直接成功
			if order.Status == model.OrderStatusPaid {
				return nil
			} else if order.Status != model.OrderStatusPending {
				return ErrRecordNotFound
			}

			// 2. Pending -> Paid 并更新支付时间
			now := time.Now()
			result := tx.Model(&model.Order{}).
				Where("id = ? AND status = ? AND expire_at >= ? AND stock_rollback_status = ? AND deleted_at IS NULL",
					orderID, model.OrderStatusPending, now, model.StockRollbackStatusPending).
				Updates(map[string]interface{}{
					"status":  model.OrderStatusPaid,
					"paid_at": now,
				})

			if result.Error != nil {
				return ErrServerInternal
			} else if result.RowsAffected == 0 {
				return ErrRecordNotFound
			}

			// 3. 扣减正式库存
			result = tx.Model(&model.Gift{}).
				Where("id = ? AND count > 0 AND deleted_at IS NULL", order.GiftID).
				Update("count", gorm.Expr("count - ?", 1))

			if result.Error != nil {
				return ErrServerInternal
			} else if result.RowsAffected == 0 {
				return ErrRecordNotFound
			}

			return nil
		})

	return err
}

// CancelTempOrder 用户取消订单
func (dao *gormOrderDAO) CancelTempOrder(ctx context.Context, orderID int64) error {
	// 1. Pending -> Cancelled
	now := time.Now()
	result := dao.db.WithContext(ctx).Model(&model.Order{}).
		Where("id = ? AND status = ? AND expire_at >= ? AND stock_rollback_status = ? AND deleted_at IS NULL",
			orderID, model.OrderStatusPending, now, model.StockRollbackStatusPending).
		Updates(map[string]interface{}{
			"status": model.OrderStatusCancelled,
		})

	if result.Error != nil {
		return ErrServerInternal
	} else if result.RowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}
