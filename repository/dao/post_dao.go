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

// GormPostDAO 用 Gorm 实现 PostDAO
type GormPostDAO struct {
	db *gorm.DB
}

// NewPostDAO 构造函数
func NewPostDAO(db *gorm.DB) *GormPostDAO {
	return &GormPostDAO{db: db}
}

// Create 创建 Post
func (dao *GormPostDAO) Create(ctx context.Context, post *model.Post) (*model.Post, error) {
	// 0. 兜底
	if post.ID == 0 {
		post.ID = snowflake.NextID()
	}
	if post.UserID == 0 || post.Title == "" || post.Content == "" {
		return nil, ErrParamsInvalid
	}

	// 1. 操作数据库
	result := dao.db.WithContext(ctx).Create(post)
	if result.Error != nil {
		// 业务层面错误
		var mysqlErr *mysql.MySQLError
		if errors.As(result.Error, &mysqlErr) && mysqlErr.Number == 1062 {
			return nil, ErrUniqueKeyConflict
		}

		// 系统层面错误
		slog.Error(CreateFailed, "post_id", post.ID, "error", result.Error)
		return nil, ErrInternal
	}

	// 2. 返回结果
	return post, nil
}

// Delete 删除 Post
func (dao *GormPostDAO) Delete(ctx context.Context, id int64) error {
	// 1. 操作数据库
	now := time.Now()
	result := dao.db.WithContext(ctx).Model(&model.Post{}).Where("id = ? AND deleted_at IS NULL", id).Update("deleted_at", &now)
	if result.Error != nil {
		// 系统层面错误
		slog.Error(DeleteFailed, "id", id, "error", result.Error)
		return ErrInternal
	} else if result.RowsAffected == 0 {
		// 业务层面错误
		return ErrRecordNotFound
	}

	// 2. 返回结果
	return nil
}

// UpdateCount 更新 Post 的 ViewCount / CommentCount / LikeCount
func (dao *GormPostDAO) UpdateCount(ctx context.Context, id int64, field model.PostCntField, delta int) error {
	// 1. 获取更新列名
	col, err := field.Column()
	if err != nil {
		return ErrParamsInvalid
	}

	// 2. 操作数据库
	result := dao.db.WithContext(ctx).Model(&model.Post{}).Where("id = ? AND deleted_at IS NULL", id).UpdateColumn(col, gorm.Expr(col+" + ?", delta))
	if result.Error != nil {
		// 系统层面错误
		slog.Error(UpdateFailed, "id", id, "field", col, "error", result.Error)
		return ErrInternal
	}
	if result.RowsAffected == 0 {
		// 业务层面错误
		var cnt int64
		result2 := dao.db.WithContext(ctx).Model(&model.Post{}).Where("id = ? AND deleted_at IS NULL", id).Count(&cnt)
		if result2.Error != nil {
			// 系统层面错误
			slog.Error(FindFailed, "id", id, "error", result2.Error)
			return ErrInternal
		}

		if cnt == 0 {
			// 记录不存在
			return ErrRecordNotFound
		}
	}

	// 3. 返回结果
	return nil
}

// Update 更新 Post 多个字段
func (dao *GormPostDAO) Update(ctx context.Context, id int64, updates map[string]any) error {
	// 1. 操作数据库
	result := dao.db.WithContext(ctx).Model(&model.Post{}).Where("id = ? AND deleted_at IS NULL", id).Updates(updates)
	if result.Error != nil {
		// 系统层面错误
		slog.Error(UpdateFailed, "id", id, "updates", updates, "error", result.Error)
		return ErrInternal
	}
	if result.RowsAffected == 0 {
		// 业务层面错误
		var cnt int64
		result2 := dao.db.WithContext(ctx).Model(&model.Post{}).Where("id = ? AND deleted_at IS NULL", id).Count(&cnt)
		if result2.Error != nil {
			// 系统层面错误
			slog.Error(FindFailed, "id", id, "error", result2.Error)
			return ErrInternal
		}

		if cnt == 0 {
			// 记录不存在
			return ErrRecordNotFound
		}
	}

	// 2. 返回结果
	return nil
}

// GetByID 根据 Post 的 ID 查找 Post
func (dao *GormPostDAO) GetByID(ctx context.Context, id int64) (*model.Post, error) {
	// 1. 构造结构体对象
	post := &model.Post{}

	// 2. 操作数据库
	result := dao.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(post)
	if result.Error != nil {
		// 业务层面错误
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		// 系统层面错误
		slog.Error(FindFailed, "id", id, "error", result.Error)
		return nil, ErrInternal
	}

	// 3. 返回结果
	return post, nil
}

// GetByUid 根据 UserID 查找 Post
func (dao *GormPostDAO) GetByUid(ctx context.Context, id int64, pageNo, pageSize int) (int64, []*model.Post, error) {
	// 0. 兜底
	if pageNo < 1 || pageSize <= 0 || pageSize > 100 {
		return 0, nil, ErrParamsInvalid
	}

	// 1. 操作数据库
	base := dao.db.WithContext(ctx).Model(&model.Post{}).Where("user_id = ? AND deleted_at IS NULL", id)

	// 2. 获取总数
	var total int64
	result := base.Count(&total)
	if result.Error != nil {
		// 系统层面错误
		slog.Error(FindFailed, "user_id", id, "error", result.Error)
		return 0, nil, ErrInternal
	} else if total == 0 {
		// 没有帖子
		return 0, []*model.Post{}, nil
	}

	// 3. 获取帖子
	var posts []*model.Post
	offset := (pageNo - 1) * pageSize // 计算偏移量
	result = base.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&posts)
	if result.Error != nil {
		// 系统层面错误
		slog.Error(FindFailed, "user_id", id, "page_no", pageNo, "page_size", pageSize, "error", result.Error)
		return 0, nil, ErrInternal
	}

	// 4. 返回结果
	return total, posts, nil
}

// GetByPage 按页查找 Post
func (dao *GormPostDAO) GetByPage(ctx context.Context, pageNo, pageSize int) (int64, []*model.Post, error) {
	// 0. 兜底
	if pageNo < 1 || pageSize <= 0 || pageSize > 100 {
		return 0, nil, ErrParamsInvalid
	}

	// 1. 操作数据库
	base := dao.db.WithContext(ctx).Model(&model.Post{}).Where("deleted_at IS NULL")

	// 2. 获取总数
	var total int64
	result := base.Count(&total)
	if result.Error != nil {
		// 系统层面错误
		slog.Error(FindFailed, "pageNo", pageNo, "pageSize", pageSize, "error", result.Error)
		return 0, nil, ErrInternal
	} else if total == 0 {
		return 0, []*model.Post{}, nil
	}

	// 3. 获取帖子
	var posts []*model.Post
	offset := (pageNo - 1) * pageSize
	result = base.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&posts)
	if result.Error != nil {
		// 系统层面错误
		slog.Error(FindFailed, "pageNo", pageNo, "pageSize", pageSize, "error", result.Error)
		return 0, nil, ErrInternal
	}

	// 4. 返回结果
	return total, posts, nil
}

// GetByPageAndTag 根据 TagID 按页查找 Post
func (dao *GormPostDAO) GetByPageAndTag(ctx context.Context, tid int64, pageNo, pageSize int) (int64, []*model.Post, error) {
	// 0. 兜底
	if pageNo < 1 || pageSize <= 0 || pageSize > 100 {
		return 0, nil, ErrParamsInvalid
	}

	// 1. 操作数据库
	base := dao.db.WithContext(ctx).Table("posts p").Select("p.*").
		Joins("JOIN post_tag pt ON p.id = pt.post_id").Where("pt.tag_id = ? AND p.deleted_at IS NULL", tid)

	// 2. 获取总数
	var total int64
	result := base.Count(&total)
	if result.Error != nil {
		// 系统层面错误
		slog.Error(FindFailed, "tag_id", tid, "pageNo", pageNo, "pageSize", pageSize, "error", result.Error)
		return 0, nil, ErrInternal
	} else if total == 0 {
		return 0, []*model.Post{}, nil
	}

	// 3. 获取帖子
	var posts []*model.Post
	offset := (pageNo - 1) * pageSize
	result = base.Order("p.created_at DESC").Offset(offset).Limit(pageSize).Find(&posts)
	if result.Error != nil {
		// 系统层面错误
		slog.Error(FindFailed, "tag_id", tid, "pageNo", pageNo, "pageSize", pageSize, "error", result.Error)
		return 0, nil, ErrInternal
	}

	// 4. 返回结果
	return total, posts, nil
}
