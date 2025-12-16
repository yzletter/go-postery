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

// gormLikeDAO 用 Gorm 实现 LikeDAO
type gormLikeDAO struct {
	db *gorm.DB
}

// NewLikeDAO 构造函数
func NewLikeDAO(db *gorm.DB) LikeDAO {
	return &gormLikeDAO{db: db}
}

// Create 创建 Like
func (dao *gormLikeDAO) Create(ctx context.Context, like *model.Like) error {
	// 0. 兜底
	if like == nil || like.UserID == 0 || like.PostID == 0 {
		return ErrParamsInvalid
	}

	// 1. 恢复软删除
	result := dao.db.WithContext(ctx).Model(&model.Like{}).Where("user_id = ? AND post_id = ? AND deleted_at IS NOT NULL", like.UserID, like.PostID).Update("deleted_at", nil)
	if result.Error != nil {
		// 系统层面错误
		slog.Error(UpdateFailed, "like", like, "error", result.Error)
		return ErrServerInternal
	}
	if result.RowsAffected > 0 {
		// 恢复成功
		return nil
	}

	// 2. 创建新记录
	result = dao.db.WithContext(ctx).Create(like)
	if result.Error != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(result.Error, &mysqlErr) && mysqlErr.Number == 1062 { // 记录没有被软删且已存在 -> 已经点赞
			// 幂等
			return nil
		}

		// 系统层面错误
		slog.Error(CreateFailed, "like", like, "error", result.Error)
		return ErrServerInternal
	}

	return nil
}

// Delete 删除 Like
func (dao *gormLikeDAO) Delete(ctx context.Context, uid, pid int64) error {
	now := time.Now()
	result := dao.db.WithContext(ctx).Model(&model.Like{}).Where("user_id = ? AND post_id = ? AND deleted_at IS NULL", uid, pid).Update("deleted_at", &now)
	if result.Error != nil {
		// 系统层面错误
		slog.Error(DeleteFailed, "user_id", uid, "post_id", pid, "error", result.Error)
		return ErrServerInternal
	}
	if result.RowsAffected == 0 {
		// 幂等
		return nil
	}

	return nil
}

// Exists 判断 Like 存在
func (dao *gormLikeDAO) Exists(ctx context.Context, uid, pid int64) (bool, error) {
	userLike := model.Like{}
	result := dao.db.WithContext(ctx).Where("user_id = ? AND post_id = ? AND deleted_at IS NULL", uid, pid).First(&userLike)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// 业务层面错误
			return false, ErrRecordNotFound
		}

		// 系统层面错误
		slog.Error(FindFailed, "user_id", uid, "post_id", pid, "error", result.Error)
		return false, ErrServerInternal
	}
	return true, nil
}
