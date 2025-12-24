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

func (dao *gormMessageDAO) GetByIDAndTargetID(ctx context.Context, id, targetID int64) ([]*model.Message, error) {
	var messages []*model.Message

	result := dao.db.WithContext(ctx).Model(&model.Message{}).
		Where("message_from = ? AND message_to = ? AND deleted_at IS NULL", id, targetID).
		Or("message_from = ? AND message_to = ? AND deleted_at IS NULL", targetID, id).Order("created_at DESC").Find(&messages)
	if result.Error != nil {
		// 系统层面错误
		slog.Error(FindFailed, "message_from", id, "message_to", targetID, "error", result.Error)
		return nil, ErrServerInternal
	}

	return messages, nil
}

func (dao *gormMessageDAO) GetByPage(ctx context.Context, id int64, targetID int64, pageNo, pageSize int) (int64, []*model.Message, error) {
	var messages []*model.Message
	var total int64
	base := dao.db.WithContext(ctx).Model(&model.Message{}).Where("message_from = ? AND message_to = ? AND deleted_at IS NULL", id, targetID).
		Or("message_from = ? AND message_to = ? AND deleted_at IS NULL", targetID, id)

	// 查找总数
	result := base.Count(&total)
	if result.Error != nil {
		// 系统层面错误
		slog.Error(FindFailed, "message_from", id, "message_to", id, "pageNo", pageNo, "pageSize", pageSize, "error", result.Error)
		return 0, nil, ErrServerInternal
	} else if total == 0 {
		return 0, messages, nil
	}

	// 查找记录
	offset := (pageNo - 1) * pageSize
	result.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&messages)
	if result.Error != nil {
		// 系统层面错误
		slog.Error(FindFailed, "message_from", id, "message_to", id, "pageNo", pageNo, "pageSize", pageSize, "error", result.Error)
		return 0, nil, ErrServerInternal
	}

	return total, messages, nil
}
