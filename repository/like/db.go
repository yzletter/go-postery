package repository

import (
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/yzletter/go-postery/model"
	repository "github.com/yzletter/go-postery/repository/user"
	"gorm.io/gorm"
)

var (
	ErrRecordHasExist = errors.New("该记录已存在")
	ErrRecordNotExist = errors.New("该记录不存在")
)

type UserLikeDBRepository struct {
	db *gorm.DB
}

func NewUserLikeDBRepository(db *gorm.DB) *UserLikeDBRepository {
	return &UserLikeDBRepository{db: db}
}

func (repo *UserLikeDBRepository) Create(uid, pid int) error {
	userLike := model.UserLike{
		UserId: uid,
		PostId: pid,
	}
	tx := repo.db.Create(&userLike)

	// 创建成功
	if tx.Error == nil {
		return nil
	}

	var mysqlErr *mysql.MySQLError
	if errors.As(tx.Error, &mysqlErr) && mysqlErr.Number == 1062 {
		// Unique Key 冲突, 说明记录已经存在, 需要判断记录是否被软删除
		tx = repo.db.Model(&model.UserLike{}).Where("user_id = ? and post_id = ?", uid, pid).Update("delete_time", nil)
		if tx.Error != nil {
			return repository.ErrMySQLInternal
		}

		if tx.RowsAffected == 1 { // 被软删除了, 恢复记录
			return nil
		}

		return ErrRecordHasExist
	}

	// 其他内部错误
	return repository.ErrMySQLInternal
}

func (repo *UserLikeDBRepository) Delete(uid, pid int) error {
	var userLike model.UserLike
	tx := repo.db.Model(&userLike).Where("user_id = ? and post_id = ? and delete_time is null", uid, pid).Update("delete_time", time.Now())
	if tx.Error != nil {
		return repository.ErrMySQLInternal
	}
	if tx.RowsAffected == 0 {
		return ErrRecordNotExist
	}

	return nil
}
