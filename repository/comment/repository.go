package repository

import (
	"errors"
	"log/slog"
	"time"

	"github.com/yzletter/go-postery/infra/model"
	"gorm.io/gorm"
)

type CommentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) *CommentRepository {
	return &CommentRepository{
		db: db,
	}
}

func (repo *CommentRepository) Create(pid int, uid int, parentId int, content string) (int, error) {
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
		return 0, errors.New("评论发表失败")
	}
	return comment.Id, nil
}
