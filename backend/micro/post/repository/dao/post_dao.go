package dao

import (
	"context"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/yzletter/go-postery/backend/event"
	"github.com/yzletter/go-postery/backend/micro/post/model"
	"gorm.io/gorm"
)

// gormPostDAO 用 Gorm 实现 PostDAO
type gormPostDAO struct {
	db *gorm.DB
}

// NewPostDAO 构造函数
func NewPostDAO(db *gorm.DB) PostDAO {
	return &gormPostDAO{db: db}
}

// Create 创建 Post, 同时绑定 Tag 并写 Outbox
func (dao *gormPostDAO) Create(ctx context.Context, post *model.Post, tags []*model.Tag, postTags []*model.PostTag, events []*event.OutboxEvent) error {
	// 0. 兜底
	if post.ID == 0 || post.UserID == 0 || post.Title == "" || post.Content == "" {
		return ErrParamsInvalid
	}
	if len(tags) != len(postTags) {
		return ErrParamsInvalid
	}

	// 1. 操作数据库
	err := dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(post).Error; err != nil {
			return err
		}

		// 创建 Tag 并绑定 Post
		for i, tag := range tags {
			if tag == nil || postTags[i] == nil {
				return ErrParamsInvalid
			}

			var storedTag model.Tag
			// 恢复软删除的 Tag
			result := tx.Model(&model.Tag{}).Where("(name = ? OR slug = ?) AND deleted_at IS NOT NULL", tag.Name, tag.Slug).Update("deleted_at", nil)
			if result.Error != nil {
				return result.Error
			}

			// 查找或创建 Tag
			result = tx.Where("(name = ? OR slug = ?) AND deleted_at IS NULL", tag.Name, tag.Slug).First(&storedTag)
			if result.Error != nil {
				if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
					return result.Error
				}

				if err := tx.Create(tag).Error; err != nil {
					var mysqlErr *mysql.MySQLError
					if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
						if err = tx.Where("(name = ? OR slug = ?) AND deleted_at IS NULL", tag.Name, tag.Slug).First(&storedTag).Error; err != nil {
							return err
						}
					} else {
						return err
					}
				} else {
					storedTag = *tag
				}
			}

			postTag := postTags[i]
			postTag.PostID = post.ID
			postTag.TagID = storedTag.ID

			// 恢复软删除的 PostTag
			result = tx.Model(&model.PostTag{}).Where("post_id = ? AND tag_id = ? AND deleted_at IS NOT NULL", postTag.PostID, postTag.TagID).Update("deleted_at", nil)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected != 0 {
				continue
			}

			if err := tx.Create(postTag).Error; err != nil {
				var mysqlErr *mysql.MySQLError
				if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
					continue
				}
				return err
			}
		}

		// 写 Outbox
		if len(events) != 0 {
			if err := tx.Create(events).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		if errors.Is(err, ErrParamsInvalid) {
			return ErrParamsInvalid
		}
		// 业务层面错误
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return ErrUniqueKey
		}

		return ErrServerInternal
	}

	// 2. 返回结果
	return nil
}

// Delete 删除 Post 并写 Outbox
func (dao *gormPostDAO) Delete(ctx context.Context, id int64, events []*event.OutboxEvent) error {
	// 1. 操作数据库
	err := dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		result := tx.Model(&model.Post{}).Where("id = ? AND deleted_at IS NULL", id).Update("deleted_at", &now)
		if result.Error != nil {
			return result.Error
		} else if result.RowsAffected == 0 {
			// 业务层面错误
			return ErrRecordNotFound
		}

		// 写 Outbox
		if len(events) != 0 {
			if err := tx.Create(events).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		if errors.Is(err, ErrRecordNotFound) {
			return ErrRecordNotFound
		}
		return ErrServerInternal
	}

	// 2. 返回结果
	return nil
}

// UpdateCount 更新 Post 的 ViewCount / CommentCount / LikeCount
func (dao *gormPostDAO) UpdateCount(ctx context.Context, id int64, field model.PostCntField, delta int) error {
	// 1. 获取更新列名
	col, err := field.Column()
	if err != nil {
		return ErrParamsInvalid
	}

	// 2. 操作数据库
	result := dao.db.WithContext(ctx).Model(&model.Post{}).Where("id = ? AND deleted_at IS NULL", id).UpdateColumn(col, gorm.Expr(col+" + ?", delta))
	if result.Error != nil {
		return ErrServerInternal
	}
	if result.RowsAffected == 0 {
		// 业务层面错误
		var cnt int64
		result2 := dao.db.WithContext(ctx).Model(&model.Post{}).Where("id = ? AND deleted_at IS NULL", id).Count(&cnt)
		if result2.Error != nil {
			return ErrServerInternal
		}

		if cnt == 0 {
			// 记录不存在
			return ErrRecordNotFound
		}
	}

	// 3. 返回结果
	return nil
}

// Update 更新 Post 和 Tag 并写 Outbox
func (dao *gormPostDAO) Update(ctx context.Context, post *model.Post, tags []*model.Tag, postTags []*model.PostTag, events []*event.OutboxEvent) error {
	// 0. 兜底
	if post.ID == 0 || len(tags) != len(postTags) {
		return ErrParamsInvalid
	}

	// 1. 操作数据库
	err := dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 更新 Post
		updates := map[string]any{
			"title":   post.Title,
			"content": post.Content,
		}

		result := tx.Model(&model.Post{}).Where("id = ? AND deleted_at IS NULL", post.ID).Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			// 业务层面错误
			var cnt int64
			result2 := tx.Model(&model.Post{}).Where("id = ? AND deleted_at IS NULL", post.ID).Count(&cnt)
			if result2.Error != nil {
				return result2.Error
			}

			if cnt == 0 {
				// 记录不存在
				return ErrRecordNotFound
			}
		}

		// 查询旧 Tag
		var oldTags []*model.Tag
		result = tx.Table("post_tag pt").
			Joins("JOIN tags t ON t.id = pt.tag_id").
			Select("t.id, t.name, t.slug").
			Where("pt.post_id = ? AND pt.deleted_at IS NULL AND t.deleted_at IS NULL", post.ID).
			Find(&oldTags)
		if result.Error != nil {
			return result.Error
		}

		hashBefore := make(map[string]*model.Tag)
		for _, tag := range oldTags {
			hashBefore[tag.Name] = tag
		}

		hashNow := make(map[string]struct{})
		for _, tag := range tags {
			if tag == nil {
				return ErrParamsInvalid
			}
			hashNow[tag.Name] = struct{}{}
		}

		// 删除旧绑定
		now := time.Now()
		for _, tag := range oldTags {
			if _, ok := hashNow[tag.Name]; !ok {
				result = tx.Model(&model.PostTag{}).
					Where("post_id = ? AND tag_id = ? AND deleted_at IS NULL", post.ID, tag.ID).
					Update("deleted_at", &now)
				if result.Error != nil {
					return result.Error
				}
			}
		}

		// 绑定新增 Tag
		for i, tag := range tags {
			if postTags[i] == nil {
				return ErrParamsInvalid
			}
			if _, ok := hashBefore[tag.Name]; ok {
				continue
			}

			var storedTag model.Tag
			// 恢复软删除的 Tag
			result = tx.Model(&model.Tag{}).
				Where("(name = ? OR slug = ?) AND deleted_at IS NOT NULL", tag.Name, tag.Slug).
				Update("deleted_at", nil)
			if result.Error != nil {
				return result.Error
			}

			// 查找或创建 Tag
			result = tx.Where("(name = ? OR slug = ?) AND deleted_at IS NULL", tag.Name, tag.Slug).First(&storedTag)
			if result.Error != nil {
				if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
					return result.Error
				}

				if err := tx.Create(tag).Error; err != nil {
					var mysqlErr *mysql.MySQLError
					if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
						if err = tx.Where("(name = ? OR slug = ?) AND deleted_at IS NULL", tag.Name, tag.Slug).
							First(&storedTag).Error; err != nil {
							return err
						}
					} else {
						return err
					}
				} else {
					storedTag = *tag
				}
			}

			postTag := postTags[i]
			postTag.PostID = post.ID
			postTag.TagID = storedTag.ID

			// 恢复软删除的 PostTag
			result = tx.Model(&model.PostTag{}).Where("post_id = ? AND tag_id = ? AND deleted_at IS NOT NULL", postTag.PostID, postTag.TagID).Update("deleted_at", nil)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected != 0 {
				continue
			}

			if err := tx.Create(postTag).Error; err != nil {
				var mysqlErr *mysql.MySQLError
				if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
					continue
				}
				return err
			}
		}

		// 写 Outbox
		if len(events) != 0 {
			if err := tx.Create(events).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		if errors.Is(err, ErrRecordNotFound) {
			return ErrRecordNotFound
		} else if errors.Is(err, ErrParamsInvalid) {
			return ErrParamsInvalid
		}

		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return ErrUniqueKey
		}

		return ErrServerInternal
	}

	// 2. 返回结果
	return nil
}

// GetByID 根据 Post 的 ID 查找 Post
func (dao *gormPostDAO) GetByID(ctx context.Context, id int64) (*model.Post, error) {
	// 1. 构造结构体对象
	post := &model.Post{}

	// 2. 操作数据库
	result := dao.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(post)
	if result.Error != nil {
		// 业务层面错误
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, ErrServerInternal
	}

	// 3. 返回结果
	return post, nil
}

// GetPostByTime 根据时间查找帖子 ID
func (dao *gormPostDAO) GetPostByTime(ctx context.Context, timeAt time.Time) ([]int64, error) {
	var ids []int64
	result := dao.db.WithContext(ctx).Model(&model.Post{}).
		Where("created_at >= ? AND deleted_at IS NULL", timeAt).
		Pluck("id", &ids)
	if result.Error != nil {
		return nil, ErrServerInternal
	}
	return ids, nil
}

// GetByUid 根据 UserID 查找 Post
func (dao *gormPostDAO) GetByUid(ctx context.Context, id int64, pageNo, pageSize int) (int64, []*model.Post, error) {
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
		return 0, nil, ErrServerInternal
	} else if total == 0 {
		// 没有帖子
		return 0, []*model.Post{}, nil
	}

	// 3. 获取帖子
	var posts []*model.Post
	offset := (pageNo - 1) * pageSize // 计算偏移量
	result = base.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&posts)
	if result.Error != nil {
		return 0, nil, ErrServerInternal
	}

	// 4. 返回结果
	return total, posts, nil
}

// GetByPage 按页查找 Post
func (dao *gormPostDAO) GetByPage(ctx context.Context, pageNo, pageSize int) (int64, []*model.Post, error) {
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
		return 0, nil, ErrServerInternal
	} else if total == 0 {
		return 0, []*model.Post{}, nil
	}

	// 3. 获取帖子
	var posts []*model.Post
	offset := (pageNo - 1) * pageSize
	result = base.Order("updated_at DESC").Offset(offset).Limit(pageSize).Find(&posts)
	if result.Error != nil {
		return 0, nil, ErrServerInternal
	}

	// 4. 返回结果
	return total, posts, nil
}

// GetByPageAndTag 根据 TagID 按页查找 Post
func (dao *gormPostDAO) GetByPageAndTag(ctx context.Context, tid int64, pageNo, pageSize int) (int64, []*model.Post, error) {
	// 0. 兜底
	if pageNo < 1 || pageSize <= 0 || pageSize > 100 {
		return 0, nil, ErrParamsInvalid
	}

	// 1. 操作数据库
	base := dao.db.WithContext(ctx).Table("posts p").
		Joins("JOIN post_tag pt ON p.id = pt.post_id").Where("pt.tag_id = ? AND p.deleted_at IS NULL", tid)

	// 2. 获取总数
	var total int64
	result := base.Count(&total)
	if result.Error != nil {
		return 0, nil, ErrServerInternal
	} else if total == 0 {
		return 0, []*model.Post{}, nil
	}

	// 3. 获取帖子
	var posts []*model.Post
	offset := (pageNo - 1) * pageSize
	result = base.Order("p.created_at DESC").Offset(offset).Limit(pageSize).Find(&posts)
	if result.Error != nil {
		return 0, nil, ErrServerInternal
	}

	// 4. 返回结果
	return total, posts, nil
}
