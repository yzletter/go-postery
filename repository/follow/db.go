package repository

import (
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/yzletter/go-postery/infra/snowflake"
	"github.com/yzletter/go-postery/model"
	repository "github.com/yzletter/go-postery/repository/user"
	"gorm.io/gorm"
)

type FollowDBRepository struct {
	db *gorm.DB
}

func NewFollowDBRepository(db *gorm.DB) *FollowDBRepository {
	return &FollowDBRepository{db: db}
}

func (repo *FollowDBRepository) Follow(ferId, feeId int) error {
	now := time.Now()
	var follow = &model.Follow{
		Id:         snowflake.NextID(),
		FollowerId: ferId,
		FolloweeId: feeId,
		CreateTime: &now,
		UpdateTime: &now,
		DeleteTime: nil,
	}
	tx := repo.db.Create(&follow)
	if tx.Error != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(tx.Error, &mysqlErr) && mysqlErr.Number == 1062 {
			return repository.ErrUniqueKeyConflict
		}
		return repository.ErrMySQLInternal
	}
	return nil
}

// IfFollow 判断关注关系 0 表示互不关注, 1 表示 a 关注了 b, 2 表示 b 关注了 a, 3 表示互相关注
func (repo *FollowDBRepository) IfFollow(ferId, feeId int) (int, error) {
	var result1, result2 model.Follow

	tx1 := repo.db.Where("follower_id = ? AND followee_id = ? AND delete_time is null", ferId, feeId).Take(&result1)
	tx2 := repo.db.Where("follower_id = ? AND followee_id = ? AND delete_time is null", feeId, ferId).Take(&result2)

	if tx1.Error != nil && !errors.Is(tx1.Error, gorm.ErrRecordNotFound) {
		return 0, repository.ErrMySQLInternal
	}

	if tx2.Error != nil && !errors.Is(tx2.Error, gorm.ErrRecordNotFound) {
		return 0, repository.ErrMySQLInternal
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
