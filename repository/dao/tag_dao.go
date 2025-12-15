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

type gormTagDAO struct {
	db *gorm.DB
}

func NewTagDAO(db *gorm.DB) TagDAO {
	return &gormTagDAO{db: db}
}

// Create 创建 Tag
func (dao *gormTagDAO) Create(ctx context.Context, tag *model.Tag) error {
	// 1. 恢复软删除
	result := dao.db.WithContext(ctx).Model(&model.Tag{}).Where("(name = ? OR slug = ?) AND deleted_at IS NOT NULL", tag.Name, tag.Slug).Update("deleted_at", nil)
	if result.Error != nil {
		// 系统层面错误
		slog.Error(UpdateFailed, "tag", tag, "error", result.Error)
		return ErrInternal
	}
	if result.RowsAffected != 0 {
		// 恢复成功
		return nil
	}

	// 2. 创建新记录
	result = dao.db.WithContext(ctx).Create(tag)
	if result.Error != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(result.Error, &mysqlErr) && mysqlErr.Number == 1062 { // 记录没有被软删且已存在 -> 标签已存在
			// 幂等
			return nil
		}

		// 系统层面错误
		slog.Error(CreateFailed, "tag", tag, "error", result.Error)
		return ErrInternal
	}

	return nil
}

// GetBySlug 根据 Slug 查找 Tag
func (dao *gormTagDAO) GetBySlug(ctx context.Context, slug string) (*model.Tag, error) {
	tag := &model.Tag{}
	result := dao.db.WithContext(ctx).Model(&model.Tag{}).Where("slug = ? AND deleted_at IS NULL", slug).First(tag)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// 业务层面错误
			return nil, ErrRecordNotFound
		}
		// 系统层面错误
		slog.Error(FindFailed, "tag_slug", slug, "error", result.Error)
		return nil, ErrInternal
	}
	return tag, nil
}

// GetByName 根据 Name 查找 Tag
func (dao *gormTagDAO) GetByName(ctx context.Context, name string) (*model.Tag, error) {
	tag := &model.Tag{}
	result := dao.db.WithContext(ctx).Model(&model.Tag{}).Where("name = ? AND deleted_at IS NULL", name).First(tag)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// 业务层面错误
			return nil, ErrRecordNotFound
		}
		// 系统层面错误
		slog.Error(FindFailed, "tag_name", name, "error", result.Error)
		return nil, ErrInternal
	}
	return tag, nil
}

// Bind 绑定 Post 和 Tag
func (dao *gormTagDAO) Bind(ctx context.Context, postTag *model.PostTag) error {
	// 1. 恢复软删除
	result := dao.db.WithContext(ctx).Model(&model.PostTag{}).Where("post_id = ? AND tag_id = ? AND deleted_at IS NOT NULL", postTag.PostID, postTag.TagID).Update("deleted_at", nil)
	if result.Error != nil {
		// 系统层面错误
		slog.Error(UpdateFailed, "post_tag", postTag, "error", result.Error)
		return ErrInternal
	}
	if result.RowsAffected != 0 {
		// 恢复成功
		return nil
	}

	// 2. 创建新记录
	result = dao.db.WithContext(ctx).Create(postTag)
	if result.Error != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(result.Error, &mysqlErr) && mysqlErr.Number == 1062 { // 记录没有被软删且已存在 -> 标签已经存在
			// 幂等
			return nil
		}

		// 系统层面错误
		slog.Error(CreateFailed, "post_tag", postTag, "error", result.Error)
		return ErrInternal
	}

	return nil
}

// DeleteBind 删除 Post 和 Tag 绑定关系
func (dao *gormTagDAO) DeleteBind(ctx context.Context, pid, tid int64) error {
	now := time.Now()
	result := dao.db.WithContext(ctx).Model(&model.PostTag{}).Where("post_id = ? AND tag_id = ? AND deleted_at IS NULL", pid, tid).Update("deleted_at", &now)
	if result.Error != nil {
		// 系统层面错误
		slog.Error(DeleteFailed, "post_id", pid, "tag_id", tid, "error", result.Error)
		return ErrInternal
	}
	if result.RowsAffected == 0 {
		// 幂等
		return nil
	}

	return nil
}

// FindTagsByPostID 根据 PostID 查找 Tags
func (dao *gormTagDAO) FindTagsByPostID(ctx context.Context, pid int64) ([]string, error) {
	var names []string
	result := dao.db.WithContext(ctx).Table("post_tag pt").
		Joins("JOIN tags t ON t.id = pt.tag_id").
		Where("pt.post_id = ? AND pt.deleted_at IS NULL AND t.deleted_at IS NULL", pid).
		Pluck("t.name", &names)

	if result.Error != nil {
		// 系统层面错误
		slog.Error(FindFailed, "post_id", pid, "error", result.Error)
		return nil, ErrInternal
	}
	return names, nil
}
