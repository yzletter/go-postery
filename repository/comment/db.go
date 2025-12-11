package repository

import (
	"errors"
	"log/slog"
	"time"

	"github.com/yzletter/go-postery/errno"
	"github.com/yzletter/go-postery/infra/snowflake"
	"github.com/yzletter/go-postery/model"
	"gorm.io/gorm"
)

type CommentDBRepository struct {
	db *gorm.DB
}

func NewCommentDBRepository(db *gorm.DB) *CommentDBRepository {
	return &CommentDBRepository{
		db: db,
	}
}

func (repo *CommentDBRepository) Create(pid int, uid int, parentId int, replyId int, content string) (model.Comment, error) {
	now := time.Now()
	comment := model.Comment{
		Id:         snowflake.NextID(),
		PostId:     pid,      // 所属帖子 id
		ParentId:   parentId, // 父评论 id, 若为 0 则为主评论
		ReplyId:    replyId,  // 当前评论所评论的Id
		UserId:     uid,      // 评论的用户 id
		Content:    content,  // 内容
		CreateTime: &now,
		DeleteTime: nil,
	}
	if err := repo.db.Create(&comment).Error; err != nil {
		slog.Error("评论发表失败", "error", err)
		return model.Comment{}, errno.ErrCreateFailed
	}
	return comment, nil
}

func (repo *CommentDBRepository) GetByID(cid int) (model.Comment, error) {
	comment := model.Comment{Id: cid}
	// Find 不报 ErrRecordNotFound
	tx := repo.db.Select("*").Where("delete_time is null").First(&comment)
	if tx.Error != nil {
		if !errors.Is(tx.Error, gorm.ErrRecordNotFound) { // 并非未找到, 而是其他错误
			slog.Error("评论查找失败", "cid", cid, "error", tx.Error)
			return model.Comment{}, errno.ErrGetFailed
		} else {
			return model.Comment{}, errno.ErrRecordNotFound
		}
	}
	return comment, nil
}

func (repo *CommentDBRepository) Delete(cid int) error {
	tx := repo.db.Model(&model.Comment{}).Where("id = ?", cid).Or("parent_id = ?", cid).Update("delete_time", time.Now())
	if tx.Error != nil {
		slog.Error("删除失败", "cid", cid)
		return errno.ErrDeleteFailed
	} else {
		if tx.RowsAffected == 0 {
			return errno.ErrRecordNotFound
		} else {
			return nil
		}
	}
}

func (repo *CommentDBRepository) GetByPostID(pid int) ([]model.Comment, error) {
	var comments []model.Comment
	// 按时间降序
	tx := repo.db.Model(&model.Comment{}).Where("post_id = ?", pid).Where("delete_time is null").Order("create_time desc").Find(&comments)
	if tx.Error != nil {
		slog.Error("获取帖子的评论失败", "pid", pid, "error", tx.Error)
		return nil, errno.ErrGetFailed
	}

	return comments, nil
}
