package repository

import "gorm.io/gorm"

type FollowDBRepository struct {
	db *gorm.DB
}

func NewFollowDBRepository(db *gorm.DB) *FollowDBRepository {
	return &FollowDBRepository{db: db}
}
