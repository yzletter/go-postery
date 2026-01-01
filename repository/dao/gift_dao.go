package dao

import "gorm.io/gorm"

type gormGiftDAO struct {
	db *gorm.DB
}

func NewGiftDAO(db *gorm.DB) GiftDAO {
	return &gormGiftDAO{db: db}
}
