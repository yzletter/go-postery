package repository

import (
	"errors"
	"log/slog"

	"github.com/yzletter/go-postery/model"
	"gorm.io/gorm"
)

// todo 代码优化

var (
	ErrUniqueKeyConflict = errors.New("唯一键冲突")
	ErrMySQLInternal     = errors.New("数据库内部错误")
	ErrUidInvalid        = errors.New("用户 ID 错误")
)

// UserDBRepository 用 Gorm 实现 UserDBRepo
type UserDBRepository struct {
	db *gorm.DB
}

func NewUserDBRepository(db *gorm.DB) *UserDBRepository {
	return &UserDBRepository{
		db: db,
	}
}

func (repo *UserDBRepository) GetByID(uid int) (bool, model.User) {
	user := model.User{ID: int64(uid)}
	tx := repo.db.Select("*").First(&user) // 隐含的where条件是id, 注意：Find不会返回ErrRecordNotFound
	if tx.Error != nil {
		// 若错误不是记录未找到, 记录系统错误
		if !errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			slog.Error("go-postery GetBriefById : 查找用户失败", "uid", uid, "error", tx.Error)
		}
		return false, model.User{}
	}
	return true, user
}
