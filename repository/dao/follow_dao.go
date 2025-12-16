package dao

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/yzletter/go-postery/model"
	"gorm.io/gorm"
)

// gormFollowDAO 用 Gorm 实现 FollowDAO
type gormFollowDAO struct {
	db *gorm.DB
}

// NewFollowDAO 构造函数
func NewFollowDAO(db *gorm.DB) FollowDAO {
	return &gormFollowDAO{db: db}
}

// Create 创建 ferID 关注 feeID
func (dao *gormFollowDAO) Create(ctx context.Context, follow *model.Follow) error {
	// 1. 操作数据库
	result := dao.db.WithContext(ctx).Model(&model.Follow{}).Where("follower_id = ? AND followee_id = ? AND deleted_at IS NOT NULL", follow.FollowerID, follow.FolloweeID).Update("deleted_at", nil)
	if result.Error != nil {
		// 系统层面错误
		return ErrServerInternal
	}
	if result.RowsAffected > 0 {
		// 恢复软删除成功
		return nil
	}

	// 2. 新建记录
	result = dao.db.WithContext(ctx).Create(follow)
	if result.Error != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(result.Error, &mysqlErr) && mysqlErr.Number == 1062 {
			// 唯一键冲突, 说明记录已经存在, 并且没有被软删除, 幂等成功
			return nil
		}
		// 系统层面错误
		slog.Error(CreateFailed, "follower_id", follow.FollowerID, "followee_id", follow.FolloweeID, "error", result.Error)
		return ErrServerInternal
	}

	// 3. 返回结果
	return nil
}

// Delete 删除 ferID 关注 feeID
func (dao *gormFollowDAO) Delete(ctx context.Context, ferID, feeID int64) error {
	now := time.Now()
	// 1. 操作数据库
	result := dao.db.WithContext(ctx).Model(&model.Follow{}).Where("follower_id = ? AND followee_id = ? AND deleted_at IS NULL", ferID, feeID).Update("deleted_at", &now)
	if result.Error != nil {
		// 系统层面错误
		slog.Error(UpdateFailed, "follower_id", ferID, "followee_id", feeID, "error", result.Error)
		return ErrServerInternal
	} else if result.RowsAffected == 0 {
		// 幂等成功
		return nil
	}

	// 2. 返回结果
	return nil
}

// Exists 判断存在关注关系 0 表示互不关注, 1 表示 a 关注 b, 2 表示 b 关注 a, 3 表示互相关注
func (dao *gormFollowDAO) Exists(ctx context.Context, ferID, feeID int64) (int, error) {
	exists := func(a, b int64) (bool, error) {
		var cnt int64
		result := dao.db.WithContext(ctx).Model(&model.Follow{}).Where("follower_id = ? AND followee_id = ? AND deleted_at IS NULL", a, b).Count(&cnt)
		if result.Error != nil {
			slog.Error(UpdateFailed, "follower_id", ferID, "followee_id", feeID, "error", result.Error)
			return false, ErrServerInternal
		}
		return cnt > 0, nil
	}

	condition1, err := exists(ferID, feeID)
	if err != nil {
		slog.Error(UpdateFailed, "follower_id", ferID, "followee_id", feeID, "error", err.Error)
		return 0, ErrServerInternal
	}
	condition2, err := exists(feeID, ferID)
	if err != nil {
		slog.Error(UpdateFailed, "follower_id", ferID, "followee_id", feeID, "error", err.Error)
		return 0, ErrServerInternal
	}

	switch {
	// 互相关注
	case condition1 && condition2:
		return 3, nil
	// 单方面关注
	case condition1:
		return 1, nil
	case condition2:
		return 2, nil
	// 互不关注
	default:
		return 0, nil
	}
}

// GetFollowers 按页返回关注当前用户的 ID 并按时间排序
func (dao *gormFollowDAO) GetFollowers(ctx context.Context, id int64, pageNo, pageSize int) (int64, []int64, error) {
	base := dao.db.WithContext(ctx).Model(&model.Follow{}).Where("followee_id = ? AND deleted_at IS NULL", id)

	var total int64
	result := base.Count(&total)
	if result.Error != nil {
		// 系统层面错误
		slog.Error(FindFailed, "followee_id", id, "pageNo", pageNo, "pageSize", pageSize, "error", result.Error)
		return 0, nil, ErrServerInternal
	}
	if total == 0 {
		return 0, []int64{}, nil
	}

	var ids []int64
	offset := (pageNo - 1) * pageSize
	result = base.Order("created_at DESC").Offset(offset).Limit(pageSize).Pluck("follower_id", &ids)
	// Find 不会返回 RecordNotFound
	if result.Error != nil {
		slog.Error(FindFailed, "followee_id", id, "pageNo", pageNo, "pageSize", pageSize, "error", result.Error)
		return 0, nil, ErrServerInternal
	}

	// 2. 返回结果
	return total, ids, nil
}

// GetFollowees 按页返回当前用户关注的所有 ID 并按时间排序
func (dao *gormFollowDAO) GetFollowees(ctx context.Context, id int64, pageNo, pageSize int) (int64, []int64, error) {
	base := dao.db.WithContext(ctx).Model(&model.Follow{}).Where("follower_id = ? AND deleted_at IS NULL", id)

	var total int64
	result := base.Count(&total)
	if result.Error != nil {
		// 系统层面错误
		slog.Error(FindFailed, "follower_id", id, "pageNo", pageNo, "pageSize", pageSize, "error", result.Error)
		return 0, nil, ErrServerInternal
	}
	if total == 0 {
		return 0, []int64{}, nil
	}

	var ids []int64
	offset := (pageNo - 1) * pageSize
	result = base.Order("created_at DESC").Offset(offset).Limit(pageSize).Pluck("followee_id", &ids)
	// Find 不会返回 RecordNotFound
	if result.Error != nil {
		slog.Error(FindFailed, "follower_id", id, "pageNo", pageNo, "pageSize", pageSize, "error", result.Error)
		return 0, nil, ErrServerInternal
	}

	// 2. 返回结果
	return total, ids, nil
}
