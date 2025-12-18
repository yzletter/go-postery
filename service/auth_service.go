package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"
	"github.com/rs/xid"
	"github.com/yzletter/go-postery/conf"
	"github.com/yzletter/go-postery/model"
	"github.com/yzletter/go-postery/service/ports"

	"time"

	"github.com/redis/go-redis/v9"
	userdto "github.com/yzletter/go-postery/dto/user"
	"github.com/yzletter/go-postery/errno"
	"github.com/yzletter/go-postery/repository"
)

type authService struct {
	userRepo   repository.UserRepository
	jwtManager ports.JwtManager
	passHasher ports.PasswordHasher
	idGen      ports.IDGenerator
	client     redis.UniversalClient
}

// NewAuthService 构造函数
func NewAuthService(userRepo repository.UserRepository, jwtManager ports.JwtManager, passHasher ports.PasswordHasher, idGen ports.IDGenerator, client redis.UniversalClient) AuthService {
	return &authService{
		userRepo:   userRepo,
		jwtManager: jwtManager,
		passHasher: passHasher,
		idGen:      idGen,
		client:     client,
	}
}

// Register 注册
func (svc *authService) Register(ctx context.Context, username, email, password string) (userdto.BriefDTO, error) {
	var empty userdto.BriefDTO
	// 参数校验
	if username == "" || password == "" || email == "" {
		return empty, errno.ErrInvalidParam
	}
	if len(password) < 8 {
		return empty, errno.ErrPasswordWeak
	}

	// 对密码进行加密
	passwordHash, err := svc.passHasher.Hash(password)
	if err != nil {
		slog.Error("PasswordHasher Hash Failed", "error", err)
		return empty, errno.ErrServerInternal
	}

	// 构造指针
	u := &model.User{
		ID:           svc.idGen.NextID(),
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		Status:       1,
	}

	// 创建记录
	err = svc.userRepo.Create(ctx, u)
	if err != nil {
		if errors.Is(err, repository.ErrUniqueKey) {
			return empty, errno.ErrUserDuplicated
		}
		return empty, errno.ErrServerInternal
	}

	return userdto.ToBriefDTO(u), nil
}

// Login 登录
func (svc *authService) Login(ctx context.Context, username, pass string) (userdto.BriefDTO, error) {
	var empty userdto.BriefDTO

	// 参数校验
	if username == "" || pass == "" {
		return empty, errno.ErrInvalidCredential
	}

	// 获取用户
	user, err := svc.userRepo.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return empty, errno.ErrUserNotFound
		}
		return empty, errno.ErrServerInternal
	}
	if user == nil {
		return empty, errno.ErrServerInternal
	}

	// 比较密码
	err = svc.passHasher.Compare(user.PasswordHash, pass)
	if err != nil {
		if errors.Is(err, ports.ErrInvalidPassword) { // 密码错误, 返回为账号或密码错误
			return empty, errno.ErrInvalidCredential
		}
		return empty, errno.ErrServerInternal
	}

	return userdto.ToBriefDTO(user), nil
}

// ClearTokens 清除 Tokens
func (svc *authService) ClearTokens(ctx context.Context, accessToken, refreshToken string) error {
	// 删除 refreshToken
	if refreshToken != "" {
		if err := svc.client.Del(ctx, conf.RefreshTokenPrefix+refreshToken).Err(); err != nil {
			return errno.ErrLogoutFailed
		}
	}

	// 拉黑 ssid
	if accessToken != "" {
		if claim, err := svc.VerifyAccessToken(accessToken); err == nil && claim != nil && claim.SSid != "" {
			ttl := time.Duration(conf.RefreshTokenMaxAgeSecs) * time.Second
			// accessToken 解析失败就跳过，不影响 logout 成功
			_ = svc.client.Set(ctx, conf.ClearTokenPrefix+claim.SSid, "", ttl).Err()
		}
	}

	return nil
}

// IssueTokens 签发 Token
func (svc *authService) IssueTokens(ctx context.Context, id int64, role int, agent string) (string, string, error) {
	// 参数校验
	if role > 1 || role < 0 {
		role = 0
	}

	// AccessToken 的 Claims
	ssid := uuid.New().String()
	expir := time.Now().Add(conf.AccessTokenExpiration * time.Second)
	accessClaims := ports.JWTTokenClaims{
		Uid:       id,
		SSid:      ssid,
		Role:      role,
		UserAgent: agent,
		Issuer:    "go-postery",
		ExpiresAt: &expir,
	}

	// 生成 AccessToken
	accessToken, err := svc.jwtManager.GenToken(accessClaims)
	if err != nil {
		return "", "", errno.ErrServerInternal
	}

	// 生成 RefreshTokenMaxAgeSecs
	refreshToken := xid.New().String()

	// 将 < auth:refresh:xxxxxx, ssid > 存入
	mp := map[string]any{
		"user_id": id,
		"ssid":    ssid,
		"role":    role,
	}

	// 避免返回了 Token 但服务端没存的不一致
	pipe := svc.client.Pipeline()
	ttl := time.Duration(conf.RefreshTokenMaxAgeSecs) * time.Second
	pipe.HSet(ctx, conf.RefreshTokenPrefix+refreshToken, mp)
	pipe.Expire(ctx, conf.RefreshTokenPrefix+refreshToken, ttl)
	if _, err := pipe.Exec(ctx); err != nil {
		return "", "", errno.ErrServerInternal
	}

	return accessToken, refreshToken, nil
}

func (svc *authService) VerifyAccessToken(tokenString string) (*ports.JWTTokenClaims, error) {
	claim, err := svc.jwtManager.VerifyToken(tokenString)
	if err != nil {
		return nil, errno.ErrUnauthorized
	}
	return claim, nil
}
