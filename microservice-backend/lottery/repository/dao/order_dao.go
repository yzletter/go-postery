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

// RecycleTempOrder 回收超时订单，返回是否需要回补缓存库存。
func (dao *gormOrderDAO) RecycleTempOrder(ctx context.Context, uid int64, orderID int64) (bool, error) {
	now := time.Now()
	result := dao.db.WithContext(ctx).Model(&model.Order{}).
		Where("id = ? AND user_id = ? AND status = ? AND expire_at <= ? AND stock_rollback_status IN ? AND deleted_at IS NULL",
			orderID, uid, model.OrderStatusPending, now, rollbackUnfinishedStatuses()).
		Updates(map[string]interface{}{
			"status": model.OrderStatusExpired,
		})

	if result.Error != nil {
		return false, ErrServerInternal
	}
	if result.RowsAffected > 0 {
		return true, nil
	}

	order, err := dao.getByIDAndUserID(ctx, orderID, uid)
	if err != nil {
		if errors.Is(err, ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	switch order.Status {
	case model.OrderStatusExpired:
		return rollbackNeeded(order), nil
	case model.OrderStatusCancelled:
		return rollbackNeeded(order), nil
	case model.OrderStatusPaid:
		return false, nil
	case model.OrderStatusPending:
		if order.ExpireAt.After(now) {
			return false, ErrRecordNotFound
		}
		return false, ErrServerInternal
	default:
		return false, nil
	}
}

func (dao *gormOrderDAO) MarkRollbackDone(ctx context.Context, orderID int64) error {
	result := dao.db.WithContext(ctx).Model(&model.Order{}).
		Where("id = ? AND status IN ? AND stock_rollback_status IN ? AND deleted_at IS NULL",
			orderID, rollbackOrderStatuses(), rollbackUnfinishedStatuses()).
		Updates(map[string]interface{}{
			"stock_rollback_status": model.StockRollbackStatusDone,
			"next_rollback_at":      nil,
		})

	if result.Error != nil {
		return ErrServerInternal
	}
	if result.RowsAffected > 0 {
		return nil
	}

	order, err := dao.getByID(ctx, orderID)
	if err != nil {
		return err
	}
	if isRollbackOrderStatus(order.Status) && order.StockRollbackStatus == model.StockRollbackStatusDone {
		return nil
	}
	return ErrRecordNotFound
}

func (dao *gormOrderDAO) MarkRollbackFailed(ctx context.Context, orderID int64, nextRollbackAt time.Time) error {
	result := dao.db.WithContext(ctx).Model(&model.Order{}).
		Where("id = ? AND status IN ? AND stock_rollback_status IN ? AND deleted_at IS NULL",
			orderID, rollbackOrderStatuses(), rollbackUnfinishedStatuses()).
		Updates(map[string]interface{}{
			"stock_rollback_status":      model.StockRollbackStatusFailed,
			"stock_rollback_retry_count": gorm.Expr("stock_rollback_retry_count + ?", 1),
			"next_rollback_at":           nextRollbackAt,
		})

	if result.Error != nil {
		return ErrServerInternal
	}
	if result.RowsAffected > 0 {
		return nil
	}

	order, err := dao.getByID(ctx, orderID)
	if err != nil {
		return err
	}
	if isRollbackOrderStatus(order.Status) && order.StockRollbackStatus == model.StockRollbackStatusDone {
		return nil
	}
	return ErrRecordNotFound
}

func (dao *gormOrderDAO) ListRollbackDueOrders(ctx context.Context, limit int) ([]*model.Order, error) {
	if limit <= 0 {
		limit = 100
	}

	now := time.Now()
	var orders []*model.Order
	result := dao.db.WithContext(ctx).Model(&model.Order{}).
		Where(`stock_rollback_status IN ? AND deleted_at IS NULL AND (
			(status IN ? AND (next_rollback_at IS NULL OR next_rollback_at <= ?)) OR
			(status = ? AND expire_at <= ?)
		)`,
			rollbackUnfinishedStatuses(), rollbackOrderStatuses(), now, model.OrderStatusPending, now).
		Order("expire_at ASC, next_rollback_at ASC").
		Limit(limit).
		Find(&orders)
	if result.Error != nil {
		return nil, ErrServerInternal
	}
	return orders, nil
}

func (dao *gormOrderDAO) getByID(ctx context.Context, orderID int64) (*model.Order, error) {
	var order model.Order
	result := dao.db.WithContext(ctx).Model(&model.Order{}).
		Where("id = ? AND deleted_at IS NULL", orderID).
		First(&order)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, ErrServerInternal
	}
	return &order, nil
}

func (dao *gormOrderDAO) getByIDAndUserID(ctx context.Context, orderID, uid int64) (*model.Order, error) {
	var order model.Order
	result := dao.db.WithContext(ctx).Model(&model.Order{}).
		Where("id = ? AND user_id = ? AND deleted_at IS NULL", orderID, uid).
		First(&order)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, ErrServerInternal
	}
	return &order, nil
}

func rollbackNeeded(order *model.Order) bool {
	return order.StockRollbackStatus == model.StockRollbackStatusPending ||
		order.StockRollbackStatus == model.StockRollbackStatusFailed
}

func rollbackUnfinishedStatuses() []int {
	return []int{model.StockRollbackStatusPending, model.StockRollbackStatusFailed}
}

func rollbackOrderStatuses() []int {
	return []int{model.OrderStatusExpired, model.OrderStatusCancelled}
}

func isRollbackOrderStatus(status int) bool {
	return status == model.OrderStatusExpired || status == model.OrderStatusCancelled
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

			// 2. 订单 Pending -> Paid 并更新支付时间, 库存 Pending -> Unneeded
			now := time.Now()
			result := tx.Model(&model.Order{}).
				Where("id = ? AND status = ? AND expire_at >= ? AND stock_rollback_status = ? AND deleted_at IS NULL",
					orderID, model.OrderStatusPending, now, model.StockRollbackStatusPending).
				Updates(map[string]interface{}{
					"status":                model.OrderStatusPaid,
					"paid_at":               now,
					"stock_rollback_status": model.StockRollbackStatusUnneeded,
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
