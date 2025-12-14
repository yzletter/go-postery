package repository

import (
	"github.com/yzletter/go-postery/repository/cache"
	"github.com/yzletter/go-postery/repository/dao"
	"golang.org/x/crypto/bcrypt"
)

type CachedUserRepository struct {
	dao   dao.UserDAO
	cache cache.UserCache
}

func NewCachedUserRepository(dao dao.UserDAO, cache cache.UserCache) *CachedUserRepository {
	return &CachedUserRepository{dao: dao, cache: cache}
}

func Create() {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(params.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return nil, ErrPasswordEncrypt
	}
}
