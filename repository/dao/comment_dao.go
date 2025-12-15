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

type GormCommentDAO struct {
	db *gorm.DB
}

func NewCommentDAO(db *gorm.DB) CommentDAO {
	return &GormCommentDAO{
		db: db,
	}
}

// Create 创建 Comment
func (dao *GormCommentDAO) Create(ctx context.Context, comment *model.Comment) (*model.Comment, error) {
	result := dao.db.WithContext(ctx).Create(comment)
	if result.Error != nil {
		// 业务层面错误
		var mysqlErr *mysql.MySQLError
		if errors.As(result.Error, &mysqlErr) && mysqlErr.Number == 1062 {
			return nil, ErrUniqueKeyConflict
		}
		// 系统层面错误
		slog.Error(CreateFailed, "error", result.Error)
		return nil, ErrInternal
	}

	return comment, nil
}

// GetByID 根据 Comment 的 ID 查找 Comment
func (dao *GormCommentDAO) GetByID(ctx context.Context, id int64) (*model.Comment, error) {
	comment := &model.Comment{}
	// Find 不报 ErrRecordNotFound
	result := dao.db.WithContext(ctx).Model(&model.Comment{}).Select("*").Where("id = ? AND deleted_at IS NULL", id).First(comment)
	if result.Error != nil {
		// 业务层面错误
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}

		// 系统层面错误
		slog.Error(FindFailed, "comment_id", id, "error", result.Error)
		return nil, ErrInternal
	}

	return comment, nil
}

// Delete 软删除 Comment 并返回删除的条数
func (dao *GormCommentDAO) Delete(ctx context.Context, id int64) (int, error) {
	now := time.Now()
	result := dao.db.WithContext(ctx).Model(&model.Comment{}).Where("(id = ? OR parent_id = ?) AND deleted_at IS NULL", id, id).Update("deleted_at", &now)
	if result.Error != nil {
		slog.Error(DeleteFailed, "comment_id", id, "error", result.Error)
		return 0, ErrInternal
	}

	return int(result.RowsAffected), nil
}

// GetByPostID 查找 Post 的一级评论
func (dao *GormCommentDAO) GetByPostID(ctx context.Context, id int64, pageNo, pageSize int) (int64, []*model.Comment, error) {
	// 0. 兜底
	if pageNo < 1 || pageSize <= 0 || pageSize > 100 {
		return 0, nil, ErrParamsInvalid
	}

	// 1. 操作数据库
	base := dao.db.WithContext(ctx).Model(&model.Comment{}).Where("post_id = ? AND parent_id = 0 AND deleted_at IS NULL", id)

	// 2. 获取总数
	var total int64
	result := base.Count(&total)
	if result.Error != nil {
		// 系统层面错误
		slog.Error(FindFailed, "post_id", id, "pageNo", pageNo, "pageSize", pageSize, "error", result.Error)
		return 0, nil, ErrInternal
	} else if total == 0 {
		return 0, []*model.Comment{}, nil
	}

	// 3. 获取评论
	var comments []*model.Comment
	offset := (pageNo - 1) * pageSize
	result = base.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&comments)
	if result.Error != nil {
		// 系统层面错误
		slog.Error(FindFailed, "post_id", id, "pageNo", pageNo, "pageSize", pageSize, "error", result.Error)
		return 0, nil, ErrInternal
	}

	// 4. 返回结果
	return total, comments, nil
}

// GetRepliesByParentID 根据 Comment 的 ID 查找 Comment 的子评论
func (dao *GormCommentDAO) GetRepliesByParentID(ctx context.Context, id int64) ([]*model.Comment, error) {
	var comments []*model.Comment
	// Find 不报 ErrRecordNotFound
	result := dao.db.WithContext(ctx).Model(&model.Comment{}).Select("*").Where("parent_id = ? AND deleted_at IS NULL", id).Find(&comments)
	if result.Error != nil {
		// 系统层面错误
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			slog.Error(FindFailed, "parent_id", id, "error", result.Error)
			return nil, ErrInternal
		}
	}

	return comments, nil
}
