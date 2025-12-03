package repository

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/yzletter/go-postery/model"
	"gorm.io/gorm"
)

type GormCommentRepository struct {
	db *gorm.DB
}

func NewGormCommentRepository(db *gorm.DB) *GormCommentRepository {
	return &GormCommentRepository{
		db: db,
	}
}

func (repo *GormCommentRepository) Create(pid int, uid int, parentId int, content string) (model.Comment, error) {
	now := time.Now()
	comment := model.Comment{
		PostId:     pid,      // 所属帖子 id
		ParentId:   parentId, // 父评论 id, 若为 0 则为主评论
		UserId:     uid,      // 评论的用户 id
		Content:    content,  // 内容
		CreateTime: &now,
		DeleteTime: nil,
	}
	if err := repo.db.Create(&comment).Error; err != nil {
		slog.Error("评论发表失败", "error", err)
		return model.Comment{}, errors.New("评论发表失败")
	}
	return comment, nil
}

func (repo *GormCommentRepository) GetByID(cid int) (model.Comment, error) {
	comment := model.Comment{Id: cid}
	// Find 不报 ErrRecordNotFound
	tx := repo.db.Select("*").Where("delete_time is null").First(&comment)
	if tx.Error != nil {
		if !errors.Is(tx.Error, gorm.ErrRecordNotFound) { // 并非未找到, 而是其他错误
			slog.Error("评论查找失败", "cid", cid, "error", tx.Error)
		}
		return model.Comment{}, errors.New("")
	}
	return comment, nil
}

func (repo *GormCommentRepository) Delete(cid int) error {
	tx := repo.db.Model(&model.Comment{}).Where("id = ?", cid).Update("delete_time", time.Now())
	if tx.Error != nil {
		slog.Error("删除失败", "cid", cid)
		return errors.New("删除失败")
	} else {
		if tx.RowsAffected == 0 {
			return fmt.Errorf("评论 %d 不存在", cid)
		} else {
			return nil
		}
	}
}

func (repo *GormCommentRepository) GetByPostID(pid int) []model.Comment {
	var comments []model.Comment
	// 按时间降序
	tx := repo.db.Model(&model.Comment{}).Where("post_id = ?", pid).Where("delete_time is null").Order("create_time desc").Find(&comments)
	if tx.Error != nil {
		slog.Error("获取帖子的评论失败", "pid", pid, "error", tx.Error)
		return nil
	}

	return comments
}
