package dao

import (
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

func (dao *GormFollowDAO) Follow(ferId, feeId int) error {
	// 先查是否有软删除的记录
	tx := dao.db.Model(&model.Follow{}).Where("follower_id = ? AND followee_id = ? AND delete_time is not null", ferId, feeId).Update("delete_time", nil)
	if tx.Error != nil {
		return ErrInternal
	}
	if tx.RowsAffected > 0 {
		return nil
	}

	// 没有软删除的记录进行创建记录
	now := time.Now()
	var follow = model.Follow{
		Id:         int(snowflake.NextID()),
		FollowerId: ferId,
		FolloweeId: feeId,
		CreateTime: &now,
		UpdateTime: &now,
		DeleteTime: nil,
	}
	tx = dao.db.Create(&follow)
	if tx.Error != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(tx.Error, &mysqlErr) && mysqlErr.Number == 1062 {
			return ErrUniqueKeyConflict
		}
		return ErrInternal
	}
	return nil
}

func (dao *GormFollowDAO) DisFollow(ferId, feeId int) error {
	tx := dao.db.Model(&model.Follow{}).Where("follower_id = ? AND followee_id = ? AND delete_time is null", ferId, feeId).Update("delete_time", time.Now())
	if tx.Error != nil {
		slog.Error("MySQL Delete Follow Failed", "error", tx.Error)
		return ErrInternal
	} else if tx.RowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

// IfFollow 判断关注关系 0 表示互不关注, 1 表示 a 关注了 b, 2 表示 b 关注了 a, 3 表示互相关注
func (dao *GormFollowDAO) IfFollow(ferId, feeId int) (int, error) {
	var result1, result2 model.Follow

	tx1 := dao.db.Where("follower_id = ? AND followee_id = ? AND delete_time is null", ferId, feeId).Take(&result1)
	tx2 := dao.db.Where("follower_id = ? AND followee_id = ? AND delete_time is null", feeId, ferId).Take(&result2)

	if tx1.Error != nil && !errors.Is(tx1.Error, gorm.ErrRecordNotFound) {
		return 0, ErrInternal
	}

	if tx2.Error != nil && !errors.Is(tx2.Error, gorm.ErrRecordNotFound) {
		return 0, ErrInternal
	}

	condition1 := tx1.Error == nil
	condition2 := tx2.Error == nil

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

// GetFollowers 返回关注当前用户的所有 id 并按时间排序
func (dao *GormFollowDAO) GetFollowers(uid int) ([]int, error) {
	var ids []int
	tx := dao.db.Model(&model.Follow{}).Where("followee_id = ? AND delete_time is null", uid).Order("create_time desc").Pluck("follower_id", &ids)
	// Find 不会返回 RecordNotFound
	if tx.Error != nil {
		return nil, ErrInternal
	}
	return ids, nil
}

// GetFollowees 返回当前用户关注的所有 id 并按时间排序
func (dao *GormFollowDAO) GetFollowees(uid int) ([]int, error) {
	var ids []int
	tx := dao.db.Model(&model.Follow{}).Where("follower_id = ? AND delete_time is null", uid).Order("create_time desc").Pluck("followee_id", &ids)
	// Find 不会返回 RecordNotFound
	if tx.Error != nil {
		return nil, ErrInternal
	}
	return ids, nil
}
