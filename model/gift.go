package model

import "time"

type Gift struct {
	ID          int64      `gorm:"primaryKey"`
	Name        string     `gorm:"name"`
	Avatar      string     `gorm:"avatar"`
	Description string     `gorm:"description"`
	Prize       string     `gorm:"prize"`
	Count       int        `gorm:"count"`
	CreatedAt   time.Time  `gorm:"column:created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at"`
	DeletedAt   *time.Time `gorm:"column:deleted_at"`
}

func (g Gift) TableName() string {
	return "gifts"
}
