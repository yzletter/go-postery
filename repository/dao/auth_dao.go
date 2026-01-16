package dao

import (
	"context"

	"github.com/yzletter/go-postery/model"
	"gorm.io/gorm"
)

type gormAuthDAO struct {
	db *gorm.DB
}

func NewAuthDAO(db *gorm.DB) AuthDAO {
	return &gormAuthDAO{
		db: db,
	}
}

func (dao *gormAuthDAO) CreateUser(ctx context.Context, authIdentity *model.AuthIdentity, passwordHash *string) error {
	//TODO implement me
	panic("implement me")
}

func (dao *gormAuthDAO) GetAuthIdentity(ctx context.Context, authType int, identifier string) (*model.AuthIdentity, error) {
	//TODO implement me
	panic("implement me")
}

func (dao *gormAuthDAO) GetPasswordHash(ctx context.Context, uid int64) (string, error) {
	//TODO implement me
	panic("implement me")
}
