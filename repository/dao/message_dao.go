package dao

import (
	"context"
	"log/slog"

	"github.com/yzletter/go-postery/model"
	"gorm.io/gorm"
)

type gormMessageDAO struct {
	db *gorm.DB
}

func NewMessageDAO(db *gorm.DB) MessageDAO {
	return &gormMessageDAO{db: db}
}
func (dao *gormMessageDAO) Create(ctx context.Context, message *model.Message) error {
	result := dao.db.WithContext(ctx).Create(message)
	if result.Error != nil {
		return ErrServerInternal
	}

	return nil
}

func (dao *gormMessageDAO) GetByID(ctx context.Context, id, targetID int64) ([]*model.Message, error) {
	var messages []*model.Message

	result := dao.db.WithContext(ctx).Model(&model.Message{}).
		Where("message_from = ? AND message_to = ? AND deleted_at IS NULL", id, targetID).
		Or("message_from = ? AND message_to = ? AND deleted_at IS NULL", targetID, id).Order("created_at ASEC").Find(&messages)
	if result.Error != nil {
		// 系统层面错误
		slog.Error(FindFailed, "id", id, "target_id", targetID, "error", result.Error)
		return nil, ErrServerInternal
	}

	return messages, nil
}
