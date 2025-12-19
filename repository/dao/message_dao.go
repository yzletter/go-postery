package dao

import "gorm.io/gorm"

type gormMessageDAO struct {
	db *gorm.DB
}

func NewMessageDAO(db *gorm.DB) MessageDAO {
	return &gormMessageDAO{db: db}
}
