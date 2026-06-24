package dao

import (
	"context"

	"github.com/yzletter/go-postery/backend/micro/auth/model"
)

const (
	DeleteFailed = "MySQL DeleteScore Record Failed"
	CreateFailed = "MySQL Create Record Failed"
	FindFailed   = "MySQL Find Record Failed"
	UpdateFailed = "MySQL Update Record Failed"
)

type AuthDAO interface {
	// CreateUser 创建用户
	//
	// Parameter:
	//	- authAggregate: 用户聚合
	//
	// Return:
	//	- error: 可能返回的错误
	CreateUser(ctx context.Context, authAggregate *model.AuthAggregate) error

	// GetAuthIdentity 根据登录方式和凭证获取登录认证
	//
	// Parameter:
	//	- authType: 认证方式
	//	- identifier: 登录凭证
	//
	// Return:
	//	- *model.AuthIdentity: 登录认证
	//	- error: 可能返回的错误
	GetAuthIdentity(ctx context.Context, authType int, identifier string) (*model.AuthIdentity, error)

	// GetAuthIdentityByIdentifier 根据凭证获取登录认证
	//
	// Parameter:
	//	- identifier: 登录凭证
	//
	// Return:
	//	- *model.AuthIdentity: 登录认证
	//	- error: 可能返回的错误
	GetAuthIdentityByIdentifier(ctx context.Context, identifier string) (*model.AuthIdentity, error)

	// GetAuthIdentityByAuthType 根据认证方式获取登录认证
	//
	// Parameter:
	//	- uid: 用户 ID
	//	- authType: 认证方式
	//
	// Return:
	//	- *model.AuthIdentity: 登录认证
	//	- error: 可能返回的错误
	GetAuthIdentityByAuthType(ctx context.Context, uid int64, authType int) (*model.AuthIdentity, error)

	// GetAuthIdentityByUID 获取用户身份认证
	//
	// Parameter:
	//	- uid: 用户 ID
	//
	// Return:
	//	- []*model.AuthIdentity: 用户身份认证列表
	//	- error: 可能返回的错误
	GetAuthIdentityByUID(ctx context.Context, uid int64) ([]*model.AuthIdentity, error)

	// GetPasswordHash 根据 UID 获取用户密码
	//
	// Parameter:
	//	- uid: 用户 ID
	//
	// Return:
	//	- string: 密码哈希
	//	- error: 可能返回的错误
	GetPasswordHash(ctx context.Context, uid int64) (string, error)

	// UpdatePasswordHash 修改用户密码
	//
	// Parameter:
	//	- uid: 用户 ID
	//	- passwordHash: 密码哈希
	//
	// Return:
	//	- error: 可能返回的错误
	UpdatePasswordHash(ctx context.Context, uid int64, passwordHash string) error

	// HasPassword 查询密码状态
	//
	// Parameter:
	//	- uid: 用户 ID
	//
	// Return:
	//	- bool: 是否已设置密码
	//	- error: 可能返回的错误
	HasPassword(ctx context.Context, uid int64) (bool, error)

	// SetPassword 初始化密码
	//
	// Parameter:
	//	- authPassword: 用户密码
	//
	// Return:
	//	- error: 可能返回的错误
	SetPassword(ctx context.Context, authPassword *model.AuthPassword) error
}
