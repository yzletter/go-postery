package repository

import (
	"context"

	"github.com/yzletter/go-postery/repository/cache"
)

type authRepository struct {
	cache cache.AuthCache
}

func (repo *authRepository) GetPhoneCode(ctx context.Context, phone string) (string, error) {
	code, err := repo.cache.GetPhoneCode(ctx, phone)
	if err != nil {
		return "", err
	}
	return code, nil
}

func (repo *authRepository) GetEmailCode(ctx context.Context, email string) (string, error) {
	code, err := repo.cache.GetEmailCode(ctx, email)
	if err != nil {
		return "", err
	}
	return code, nil
}

func NewAuthRepository(authCache cache.AuthCache) AuthRepository {
	return &authRepository{
		cache: authCache,
	}
}

func (repo *authRepository) DelRefreshToken(ctx context.Context, refreshToken string) error {
	err := repo.cache.DelRefreshToken(ctx, refreshToken)
	if err != nil {
		return toRepositoryErr(err)
	}
	return nil
}

func (repo *authRepository) CheckBlackList(ctx context.Context, ssid string) (bool, error) {
	exist, err := repo.cache.CheckBlackList(ctx, ssid)
	if err != nil {
		return false, toRepositoryErr(err)
	}

	return exist, nil
}

func (repo *authRepository) GetInfoByRefreshToken(ctx context.Context, refreshToken string) (int64, int, string, error) {
	uid, role, ssid, err := repo.cache.GetInfoByRefreshToken(ctx, refreshToken)
	if err != nil {
		return 0, 0, "", toRepositoryErr(err)
	}

	return uid, role, ssid, nil
}

func (repo *authRepository) SetBlackList(ctx context.Context, ssid string) error {
	err := repo.cache.SetBlackList(ctx, ssid)
	if err != nil {
		return toRepositoryErr(err)
	}
	return nil
}

func (repo *authRepository) SetInfo(ctx context.Context, refreshToken string, mp map[string]any) error {
	err := repo.cache.SetInfo(ctx, refreshToken, mp)
	if err != nil {
		return toRepositoryErr(err)
	}
	return nil
}
