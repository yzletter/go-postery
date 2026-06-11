package repository

import (
	"context"

	"github.com/yzletter/go-postery/backend/micro/auth/model"
	"github.com/yzletter/go-postery/backend/micro/auth/repository/cache"
	"github.com/yzletter/go-postery/backend/micro/auth/repository/dao"
)

type authRepository struct {
	dao   dao.AuthDAO
	cache cache.AuthCache
}

func NewAuthRepository(authDAO dao.AuthDAO, authCache cache.AuthCache) AuthRepository {
	return &authRepository{
		dao:   authDAO,
		cache: authCache,
	}
}

// CreateUser 创建用户（包括用户最小项、用户登录认证、用户密码、用户资料、注册扩展功能）
func (repo *authRepository) CreateUser(ctx context.Context, authAggregate *model.AuthAggregate) error {
	err := repo.dao.CreateUser(ctx, authAggregate)
	if err != nil {
		return toRepositoryErr(err)
	}
	return nil
}

// GetAuthIdentity 根据登录方式和凭证获取登录认证
func (repo *authRepository) GetAuthIdentity(ctx context.Context, authType int, identifier string) (*model.AuthIdentity, error) {
	authIdentity, err := repo.dao.GetAuthIdentity(ctx, authType, identifier)
	if err != nil {
		return nil, toRepositoryErr(err)
	}
	return authIdentity, nil
}

func (repo *authRepository) GetAuthIdentityByIdentifier(ctx context.Context, identifier string) (*model.AuthIdentity, error) {
	authIdentity, err := repo.dao.GetAuthIdentityByIdentifier(ctx, identifier)
	if err != nil {
		return nil, toRepositoryErr(err)
	}
	return authIdentity, nil
}

func (repo *authRepository) GetAuthIdentityByAuthType(ctx context.Context, uid int64, authType int) (*model.AuthIdentity, error) {
	authIdentity, err := repo.dao.GetAuthIdentityByAuthType(ctx, uid, authType)
	if err != nil {
		return nil, toRepositoryErr(err)
	}
	return authIdentity, nil
}

// GetAuthIdentityByUID 获取用户身份认证
func (repo *authRepository) GetAuthIdentityByUID(ctx context.Context, uid int64) (string, string, error) {
	authIdentitys, err := repo.dao.GetAuthIdentityByUID(ctx, uid)
	if err != nil {
		return "", "", toRepositoryErr(err)
	}

	if authIdentitys == nil || len(authIdentitys) == 0 {
		return "", "", nil
	}

	var phone, email string
	for _, authIdentity := range authIdentitys {
		if authIdentity == nil {
			continue
		}
		switch authIdentity.AuthType {
		case 1:
			phone = authIdentity.Identifier
		case 2:
			email = authIdentity.Identifier
		}
	}

	return phone, email, nil
}

// GetPasswordHash 根据 UID 获取用户密码
func (repo *authRepository) GetPasswordHash(ctx context.Context, uid int64) (string, error) {
	passwordHash, err := repo.dao.GetPasswordHash(ctx, uid)
	if err != nil {
		return "", toRepositoryErr(err)
	}
	return passwordHash, nil
}

// UpdatePasswordHash 修改用户密码
func (repo *authRepository) UpdatePasswordHash(ctx context.Context, uid int64, passwordHash string) error {
	err := repo.dao.UpdatePasswordHash(ctx, uid, passwordHash)
	if err != nil {
		return toRepositoryErr(err)
	}
	return nil
}

// HasPassword 查询密码状态
func (repo *authRepository) HasPassword(ctx context.Context, uid int64) (bool, error) {
	has, err := repo.dao.HasPassword(ctx, uid)
	if err != nil {
		return false, toRepositoryErr(err)
	}
	return has, nil
}

// SetPassword 初始化密码
func (repo *authRepository) SetPassword(ctx context.Context, authPassword *model.AuthPassword) error {
	err := repo.dao.SetPassword(ctx, authPassword)
	if err != nil {
		return toRepositoryErr(err)
	}

	return nil
}

// DelRefreshToken 缓存中删除 RefreshToken
func (repo *authRepository) DelRefreshToken(ctx context.Context, refreshToken string) error {
	err := repo.cache.DelRefreshToken(ctx, refreshToken)
	if err != nil {
		return toRepositoryErr(err)
	}
	return nil
}

// SetInfo 根据 RefreshToken 在缓存中存储用户信息
func (repo *authRepository) SetInfo(ctx context.Context, refreshToken string, mp map[string]any) error {
	err := repo.cache.SetInfo(ctx, refreshToken, mp)
	if err != nil {
		return toRepositoryErr(err)
	}
	return nil
}

// GetInfoByRefreshToken 根据 RefreshToken 从缓存中读取用户信息
func (repo *authRepository) GetInfoByRefreshToken(ctx context.Context, refreshToken string) (int64, int, string, error) {
	uid, role, ssid, err := repo.cache.GetInfoByRefreshToken(ctx, refreshToken)
	if err != nil {
		return 0, 0, "", toRepositoryErr(err)
	}

	return uid, role, ssid, nil
}

// SetBlackList 拉黑 SSID
func (repo *authRepository) SetBlackList(ctx context.Context, ssid string) error {
	err := repo.cache.SetBlackList(ctx, ssid)
	if err != nil {
		return toRepositoryErr(err)
	}
	return nil
}

// CheckBlackList 查看 SSID 是否被拉黑
func (repo *authRepository) CheckBlackList(ctx context.Context, ssid string) (bool, error) {
	exist, err := repo.cache.CheckBlackList(ctx, ssid)
	if err != nil {
		return false, toRepositoryErr(err)
	}

	return exist, nil
}
