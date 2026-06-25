package dao

import (
	"context"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/yzletter/go-postery/backend/event"
	"github.com/yzletter/go-postery/backend/micro/interactive/domain"
	"github.com/yzletter/go-postery/backend/micro/interactive/model"
	"github.com/yzletter/go-postery/backend/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type gormInteractiveDAO struct {
	db    *gorm.DB
	idGen ports.IDGenerator
}

func NewInteractiveDAO(db *gorm.DB, idGen ports.IDGenerator) InteractiveDAO {
	return &gormInteractiveDAO{
		db:    db,
		idGen: idGen,
	}
}

func (dao *gormInteractiveDAO) GetPostInteractive(ctx context.Context, postID int64) (model.PostInteractive, error) {
	var postInter model.PostInteractive
	result := dao.db.WithContext(ctx).Model(&model.PostInteractive{}).Where("post_id = ?", postID).First(&postInter)
	if result.Error != nil {
		// 业务层面错误
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return model.PostInteractive{}, ErrRecordNotFound
		}
		return model.PostInteractive{}, ErrServerInternal
	}

	return postInter, nil
}

func (dao *gormInteractiveDAO) GetUserInteractive(ctx context.Context, userID int64) (model.UserInteractive, error) {
	var userInter model.UserInteractive
	result := dao.db.WithContext(ctx).Model(&model.UserInteractive{}).Where("user_id = ?", userID).First(&userInter)
	if result.Error != nil {
		// 业务层面错误
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return model.UserInteractive{}, ErrRecordNotFound
		}
		return model.UserInteractive{}, ErrServerInternal
	}

	return userInter, nil
}

func (dao *gormInteractiveDAO) IncrReadCnt(ctx context.Context, consumer string, topic string, readEventPayloads ...*event.NewReadEventPayload) error {
	if dao.idGen == nil || consumer == "" || topic == "" {
		return ErrParamsInvalid
	}
	if len(readEventPayloads) == 0 {
		return nil
	}
	mp := make(map[int64]int64) // 计数

	// 开启事务
	err := dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, readEvent := range readEventPayloads {
			result := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "consumer"}, {Name: "event_id"}},
				DoNothing: true,
			}).Create(&event.ProcessedEvent{
				ID:        dao.idGen.NextID(),
				Consumer:  consumer,
				EventID:   readEvent.ID,
				Topic:     topic,
				CreatedAt: time.Time{},
			})
			if result.Error != nil {
				return result.Error
			}

			if result.RowsAffected == 0 { // 消费过
				continue
			}
			mp[readEvent.PostID]++
		}

		for id, delta := range mp {
			err := tx.
				Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "post_id"}},
					DoUpdates: clause.Assignments(map[string]interface{}{"read_count": gorm.Expr("read_count"+" + ?", delta)})}).
				Create(&model.PostInteractive{
					ID:      dao.idGen.NextID(),
					PostID:  id,
					ReadCnt: delta,
				}).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return ErrServerInternal
	}
	return nil
}

// ChangeInteractiveCnt 根据业务类型修改 Inter 计数
func (dao *gormInteractiveDAO) ChangeInteractiveCnt(ctx context.Context, biz domain.BizType, bizID int64, delta int64, processedEvent *event.ProcessedEvent) error {
	// 新建 Inter 记录时需要雪花 ID
	if dao.idGen == nil {
		return ErrServerInternal
	}
	err := dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 写消费表
		result := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "consumer"}, {Name: "event_id"}},
			DoNothing: true,
		}).Create(processedEvent)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return nil
		}
		switch biz {
		// 阅读、点赞、评论都归属帖子 Inter
		case domain.BizRead:
			err := tx.
				Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "post_id"}},
					DoUpdates: clause.Assignments(map[string]interface{}{"read_count": gorm.Expr("read_count"+" + ?", delta)})}).
				Create(&model.PostInteractive{
					ID:      dao.idGen.NextID(),
					PostID:  bizID,
					ReadCnt: delta,
				}).Error
			if err != nil {
				return err
			}

		case domain.BizLike:
			err := tx.
				Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "post_id"}},
					DoUpdates: clause.Assignments(map[string]interface{}{"like_count": gorm.Expr("like_count"+" + ?", delta)})}).
				Create(&model.PostInteractive{
					ID:      dao.idGen.NextID(),
					PostID:  bizID,
					LikeCnt: delta,
				}).Error
			if err != nil {
				return err
			}
		case domain.BizComment:
			err := tx.
				Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "post_id"}},
					DoUpdates: clause.Assignments(map[string]interface{}{"comment_count": gorm.Expr("comment_count"+" + ?", delta)})}).
				Create(&model.PostInteractive{
					ID:         dao.idGen.NextID(),
					PostID:     bizID,
					CommentCnt: delta,
				}).Error
			if err != nil {
				return err
			}

		// 关注归属用户 Inter
		case domain.BizFollow:
			err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "user_id"}},
				DoUpdates: clause.Assignments(map[string]interface{}{"follow_count": gorm.Expr("follow_count + ?", delta)}),
			}).
				Create(&model.UserInteractive{
					ID:        dao.idGen.NextID(),
					UserID:    bizID,
					FollowCnt: delta,
				}).Error
			if err != nil {
				return err
			}

		default:
			return ErrParamsInvalid
		}
		return nil
	})

	if err != nil {
		return ErrServerInternal
	}

	return nil
}

//
//// incrPostInteractiveCnt 通过 post_id 做 upsert 并累加指定字段
//func (dao *gormInteractiveDAO) incrPostInteractiveCnt(ctx context.Context, column string, inter model.PostInteractive, delta int64) *gorm.DB {
//	return dao.db.WithContext(ctx).
//		Clauses(clause.OnConflict{
//			Columns:   []clause.Column{{Name: "post_id"}},
//			DoUpdates: clause.Assignments(map[string]interface{}{column: gorm.Expr(column+" + ?", delta)}),
//		}).
//		Create(&inter)
//}

// GetFollow 获取关注关系
func (dao *gormInteractiveDAO) GetFollow(ctx context.Context, follower, followee int64) (domain.FollowType, error) {
	exists := func(a, b int64) (bool, error) {
		var cnt int64
		result := dao.db.WithContext(ctx).
			Model(&model.Follow{}).
			Where("follower_id = ? AND followee_id = ? AND deleted_at IS NULL", a, b).
			Count(&cnt)

		if result.Error != nil {
			return false, ErrServerInternal
		}
		return cnt > 0, nil
	}

	condition1, err := exists(follower, followee)
	if err != nil {
		return domain.FollowWrong, ErrServerInternal
	}
	condition2, err := exists(followee, follower)
	if err != nil {
		return domain.FollowWrong, ErrServerInternal
	}

	switch {
	// 互相关注
	case condition1 && condition2:
		return domain.FollowMutual, nil
	// 单方面关注
	case condition1:
		return domain.FollowIFollow, nil
	case condition2:
		return domain.FollowFollowMe, nil
	// 互不关注
	default:
		return domain.FollowNone, nil
	}
}

// CreateFollow 创建 ferID 关注 feeID
func (dao *gormInteractiveDAO) CreateFollow(ctx context.Context, follow *model.Follow, events ...*event.OutboxEvent) error {
	err := dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 尝试恢复记录
		result := tx.Model(&model.Follow{}).
			Where("follower_id = ? AND followee_id = ? AND deleted_at IS NOT NULL", follow.FollowerID, follow.FolloweeID).
			Update("deleted_at", nil)
		if result.Error != nil {
			return result.Error
		}

		// 2. 新建记录
		if result.RowsAffected == 0 {
			if err := tx.Create(follow).Error; err != nil {
				return err
			}
		}

		// 3. 写 Outbox
		for _, event := range events {
			if event == nil {
				continue
			}
			if err := tx.Create(event).Error; err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			// 唯一键冲突, 说明记录已经存在, 并且没有被软删除, 幂等成功
			return ErrUniqueKey
		}
		return ErrServerInternal
	}

	return nil
}

// DelFollow 删除 ferID 关注 feeID
func (dao *gormInteractiveDAO) DelFollow(ctx context.Context, ferID, feeID int64, events ...*event.OutboxEvent) error {
	now := time.Now()
	err := dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 软删除
		result := tx.Model(&model.Follow{}).
			Where("follower_id = ? AND followee_id = ? AND deleted_at IS NULL", ferID, feeID).
			Update("deleted_at", &now)
		if result.Error != nil {
			return result.Error
		} else if result.RowsAffected == 0 {
			return ErrRecordNotFound
		}

		// 2. 写 Outbox
		for _, event := range events {
			if event == nil {
				continue
			}
			if err := tx.Create(event).Error; err != nil {
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

	return nil
}

// GetFollowers 按页返回关注当前用户的 ID 并按时间排序
func (dao *gormInteractiveDAO) GetFollowers(ctx context.Context, id int64, pageNo, pageSize int) (int64, []int64, error) {
	if pageNo < 1 || pageSize <= 0 || pageSize > 100 {
		return 0, nil, ErrParamsInvalid
	}

	base := dao.db.WithContext(ctx).Model(&model.Follow{}).Where("followee_id = ? AND deleted_at IS NULL", id)

	// 1. 取总数
	var total int64
	result := base.Count(&total)
	if result.Error != nil {
		return 0, nil, ErrServerInternal
	}
	if total == 0 {
		return 0, []int64{}, nil
	}

	// 2. 取具体
	var ids []int64
	offset := (pageNo - 1) * pageSize
	result = base.Order("created_at DESC").Offset(offset).Limit(pageSize).Pluck("follower_id", &ids)
	if result.Error != nil {
		// Find 不会返回 RecordNotFound
		return 0, nil, ErrServerInternal
	}

	// 2. 返回结果
	return total, ids, nil
}

// GetFollowees 按页返回当前用户关注的所有 ID 并按时间排序
func (dao *gormInteractiveDAO) GetFollowees(ctx context.Context, id int64, pageNo, pageSize int) (int64, []int64, error) {
	if pageNo < 1 || pageSize <= 0 || pageSize > 100 {
		return 0, nil, ErrParamsInvalid
	}

	base := dao.db.WithContext(ctx).Model(&model.Follow{}).Where("follower_id = ? AND deleted_at IS NULL", id)

	// 1. 取总数
	var total int64
	result := base.Count(&total)
	if result.Error != nil {
		return 0, nil, ErrServerInternal
	}
	if total == 0 {
		return 0, []int64{}, nil
	}

	// 2. 取具体
	var ids []int64
	offset := (pageNo - 1) * pageSize
	result = base.Order("created_at DESC").Offset(offset).Limit(pageSize).Pluck("followee_id", &ids)
	// Find 不会返回 RecordNotFound
	if result.Error != nil {
		return 0, nil, ErrServerInternal
	}

	// 2. 返回结果
	return total, ids, nil
}

// CreateComment 创建 Comment
func (dao *gormInteractiveDAO) CreateComment(ctx context.Context, comment *model.Comment, events ...*event.OutboxEvent) error {
	err := dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(comment).Error; err != nil {
			return err
		}

		// 写 Outbox
		for _, event := range events {
			if event == nil {
				continue
			}
			if err := tx.Create(event).Error; err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return ErrUniqueKey
		}
		return ErrServerInternal
	}

	return nil
}

// GetCommentByID 根据 Comment 的 ID 查找 Comment
func (dao *gormInteractiveDAO) GetCommentByID(ctx context.Context, id int64) (model.Comment, error) {
	var comment model.Comment
	// Find 不报 ErrRecordNotFound
	result := dao.db.WithContext(ctx).Model(&model.Comment{}).Where("id = ? AND deleted_at IS NULL", id).First(&comment)
	if result.Error != nil {
		// 业务层面错误
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return model.Comment{}, ErrRecordNotFound
		}

		return model.Comment{}, ErrServerInternal
	}

	return comment, nil
}

// DelComment 软删除 Comment 并返回删除的条数
func (dao *gormInteractiveDAO) DelComment(ctx context.Context, id int64, buildEvents func(cnt int) []*event.OutboxEvent) (int, error) {
	now := time.Now()
	cnt := 0
	err := dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&model.Comment{}).Where("(id = ? OR parent_id = ?) AND deleted_at IS NULL", id, id).Update("deleted_at", &now)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrRecordNotFound
		}
		cnt = int(result.RowsAffected)

		// 写 Outbox
		var events []*event.OutboxEvent
		if buildEvents != nil {
			events = buildEvents(cnt)
		}
		for _, e := range events {
			if e == nil {
				continue
			}
			if err := tx.Create(e).Error; err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		if errors.Is(err, ErrRecordNotFound) {
			return 0, ErrRecordNotFound
		}
		return 0, ErrServerInternal
	}

	return cnt, nil
}

// GetCommentByPostID 查找 Post 的一级评论
func (dao *gormInteractiveDAO) GetCommentByPostID(ctx context.Context, id int64, pageNo, pageSize int) (int64, []model.Comment, error) {
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
		return 0, nil, ErrServerInternal
	} else if total == 0 {
		return 0, []model.Comment{}, nil
	}

	// 3. 获取评论
	var comments []model.Comment
	offset := (pageNo - 1) * pageSize
	result = base.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&comments)
	if result.Error != nil {
		return 0, nil, ErrServerInternal
	}

	// 4. 返回结果
	return total, comments, nil
}

// GetCommentRepliesByParentID 根据多个 Comment 的 ID 查找 Comment 的子评论
func (dao *gormInteractiveDAO) GetCommentRepliesByParentID(ctx context.Context, id int64, pageNo, pageSize int) (int64, []model.Comment, error) {
	if pageNo < 1 || pageSize <= 0 || pageSize > 100 {
		return 0, nil, ErrParamsInvalid
	}

	var comments []model.Comment

	base := dao.db.WithContext(ctx).Model(&model.Comment{}).Where("parent_id = ? AND deleted_at is NULL", id)

	var total int64
	result := base.Count(&total)
	if result.Error != nil {
		return total, nil, ErrServerInternal
	} else if total == 0 {
		return 0, comments, nil
	}

	offset := (pageNo - 1) * pageSize
	result = base.Order("parent_id ASC, created_at ASC").Offset(offset).Limit(pageSize).Find(&comments)
	if result.Error != nil {
		return 0, nil, ErrServerInternal
	}

	return total, comments, nil
}

// CreateLike 创建 Like
func (dao *gormInteractiveDAO) CreateLike(ctx context.Context, like *model.Like, events ...*event.OutboxEvent) error {
	// 0. 兜底
	if like == nil || like.UserID == 0 || like.PostID == 0 {
		return ErrParamsInvalid
	}

	err := dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 恢复软删除
		result := tx.Model(&model.Like{}).
			Where("user_id = ? AND post_id = ? AND deleted_at IS NOT NULL", like.UserID, like.PostID).
			Update("deleted_at", nil)
		if result.Error != nil {
			return result.Error
		}

		// 2. 创建新记录
		if result.RowsAffected == 0 {
			if err := tx.Create(like).Error; err != nil {
				return err
			}
		}

		// 3. 写 Outbox
		for _, event := range events {
			if event == nil {
				continue
			}
			if err := tx.Create(event).Error; err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 { // 记录没有被软删且已存在 -> 已经点赞
			// 幂等
			return ErrUniqueKey
		}

		return ErrServerInternal
	}

	return nil
}

// DelLike 删除 Like
func (dao *gormInteractiveDAO) DelLike(ctx context.Context, uid, pid int64, events ...*event.OutboxEvent) error {
	now := time.Now()
	err := dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&model.Like{}).Where("user_id = ? AND post_id = ? AND deleted_at IS NULL", uid, pid).Update("deleted_at", &now)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrRecordNotFound
		}

		// 写 Outbox
		for _, event := range events {
			if event == nil {
				continue
			}
			if err := tx.Create(event).Error; err != nil {
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

	return nil
}

// GetLike 获取点赞关系
func (dao *gormInteractiveDAO) GetLike(ctx context.Context, uid, pid int64) (bool, error) {
	userLike := model.Like{}
	result := dao.db.WithContext(ctx).Where("user_id = ? AND post_id = ? AND deleted_at IS NULL", uid, pid).First(&userLike)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return false, ErrServerInternal
		}
		return false, nil
	}
	return true, nil
}
