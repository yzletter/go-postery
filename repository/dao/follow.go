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

type GormFollowDAO struct {
	db *gorm.DB
}

func NewFollowDAO(db *gorm.DB) FollowDAO {
	return &GormFollowDAO{db: db}
}

// Follow ferID 关注 feeID
func (dao *GormFollowDAO) Follow(ctx context.Context, ferID, feeID int64) error {
	// 1. 操作数据库
	result := dao.db.WithContext(ctx).Model(&model.Follow{}).Where("follower_id = ? AND followee_id = ? AND deleted_at IS NOT NULL", ferID, feeID).Update("deleted_at", nil)
	if result.Error != nil {
		// 系统层面错误
		return ErrInternal
	}
	if result.RowsAffected > 0 {
		// 恢复软删除成功
		return nil
	}

	// 2. 新建记录
	var follow = model.Follow{
		ID:         snowflake.NextID(),
		FollowerID: ferID,
		FolloweeID: feeID,
		DeletedAt:  nil,
	}
	result = dao.db.WithContext(ctx).Create(&follow)
	if result.Error != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(result.Error, &mysqlErr) && mysqlErr.Number == 1062 {
			// 唯一键冲突, 说明记录已经存在, 并且没有被软删除, 幂等成功
			return nil
		}
		// 系统层面错误
		slog.Error(CreateFailed, "follower_id", ferID, "followee_id", feeID, "error", result.Error)
		return ErrInternal
	}

	// 3. 返回结果
	return nil
}

// UnFollow 取消 ferID 关注 feeID
func (dao *GormFollowDAO) UnFollow(ctx context.Context, ferID, feeID int64) error {
	now := time.Now()
	// 1. 操作数据库
	result := dao.db.WithContext(ctx).Model(&model.Follow{}).Where("follower_id = ? AND followee_id = ? AND deleted_at IS NULL", ferID, feeID).Update("deleted_at", &now)
	if result.Error != nil {
		// 系统层面错误
		slog.Error(UpdateFailed, "follower_id", ferID, "followee_id", feeID, "error", result.Error)
		return ErrInternal
	} else if result.RowsAffected == 0 {
		// 幂等成功
		return nil
	}

	// 2. 返回结果
	return nil
}

// IfFollow 判断关注关系 0 表示互不关注, 1 表示 a 关注 b, 2 表示 b 关注 a, 3 表示互相关注
func (dao *GormFollowDAO) IfFollow(ctx context.Context, ferID, feeID int64) (int, error) {
	exists := func(a, b int64) (bool, error) {
		var cnt int64
		result := dao.db.WithContext(ctx).Model(&model.Follow{}).Where("follower_id = ? AND followee_id = ? AND deleted_at IS NULL", a, b).Count(&cnt)
		if result.Error != nil {
			slog.Error(UpdateFailed, "follower_id", ferID, "followee_id", feeID, "error", result.Error)
			return false, ErrInternal
		}
		return cnt > 0, nil
	}

	condition1, err := exists(ferID, feeID)
	if err != nil {
		slog.Error(UpdateFailed, "follower_id", ferID, "followee_id", feeID, "error", err.Error)
		return 0, ErrInternal
	}
	condition2, err := exists(feeID, ferID)
	if err != nil {
		slog.Error(UpdateFailed, "follower_id", ferID, "followee_id", feeID, "error", err.Error)
		return 0, ErrInternal
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
func (dao *GormFollowDAO) GetFollowers(ctx context.Context, id int64, pageNo, pageSize int) (int64, []int64, error) {
	base := dao.db.WithContext(ctx).Model(&model.Follow{}).Where("followee_id = ? AND deleted_at IS NULL", id)

	var total int64
	result := base.Count(&total)
	if result.Error != nil {
		// 系统层面错误
		slog.Error(FindFailed, "followee_id", id, "pageNo", pageNo, "pageSize", pageSize, "error", result.Error)
		return 0, nil, ErrInternal
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
		return 0, nil, ErrInternal
	}

	// 2. 返回结果
	return total, ids, nil
}

// GetFollowees 按页返回当前用户关注的所有 ID 并按时间排序
func (dao *GormFollowDAO) GetFollowees(ctx context.Context, id int64, pageNo, pageSize int) (int64, []int64, error) {
	base := dao.db.WithContext(ctx).Model(&model.Follow{}).Where("follower_id = ? AND deleted_at IS NULL", id)

	var total int64
	result := base.Count(&total)
	if result.Error != nil {
		// 系统层面错误
		slog.Error(FindFailed, "follower_id", id, "pageNo", pageNo, "pageSize", pageSize, "error", result.Error)
		return 0, nil, ErrInternal
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
		return 0, nil, ErrInternal
	}

	// 2. 返回结果
	return total, ids, nil
}
