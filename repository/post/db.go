package repository

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/yzletter/go-postery/infra/snowflake"
	"github.com/yzletter/go-postery/model"
	"gorm.io/gorm"
)

type PostDBRepository struct {
	db *gorm.DB
}

func NewPostDBRepository(db *gorm.DB) *PostDBRepository {
	return &PostDBRepository{
		db: db,
	}
}

// GetByID 根据帖子 id 获取帖子信息
func (repo *PostDBRepository) GetByID(pid int) (bool, model.Post) {
	post := model.Post{
		ID: pid,
	}
	tx := repo.db.Select("*").Where("delete_time is null").First(&post) // find 不会报 ErrNotFound
	if tx.Error != nil {
		if !errors.Is(tx.Error, gorm.ErrRecordNotFound) { // 并非未找到, 而是其他错误
			slog.Error("帖子查找失败", "pid", pid, "error", tx.Error)
		}
		return false, model.Post{}
	}

	return true, post
}

func (repo *PostDBRepository) ChangeCommentCnt(pid int, delta int) {
	var post model.Post
	tx := repo.db.Model(&post).Where("id = ?", pid).UpdateColumn("comment_count", gorm.Expr("comment_count + ?", delta))
	if tx.Error != nil {
		slog.Error("MySQL Increase Comment Count False", "error", tx.Error)
	}
}
