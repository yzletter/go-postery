package dao

import (
	"context"
	"errors"

	"github.com/yzletter/go-postery/model"
	"gorm.io/gorm"
)

type gormGiftDAO struct {
	db *gorm.DB
}

func NewGiftDAO(db *gorm.DB) GiftDAO {
	return &gormGiftDAO{db: db}
}

func (dao *gormGiftDAO) GetAll(ctx context.Context) ([]*model.Gift, error) {
	var gifts []*model.Gift
	result := dao.db.WithContext(ctx).Model(&model.Gift{}).Select("*").Find(&gifts)

	if result.Error != nil {
		return nil, ErrServerInternal
	}

	return gifts, nil
}

func (dao *gormGiftDAO) GetByID(ctx context.Context, gid int64) (*model.Gift, error) {
	var gift *model.Gift
	result := dao.db.WithContext(ctx).Model(&model.Gift{}).Where("id = ?", gid).First(&gift)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, ErrServerInternal
	}

	return gift, nil
}
