package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/rs/xid"
	"github.com/yzletter/go-postery/model"

	"time"

	"github.com/redis/go-redis/v9"
	userdto "github.com/yzletter/go-postery/dto/user"
	"github.com/yzletter/go-postery/errno"
	"github.com/yzletter/go-postery/repository"
)

const (
	FreshTokenPrefix       = "auth:refresh:"
	ClearTokenPrefix       = "auth:clear:"
	RefreshTokenExpiration = 7 * 86400 * time.Second
)

type authService struct {
	userRepo   repository.UserRepository
	jwtManager JwtManager
	passHasher PasswordHasher
	idGen      IDGenerator
	client     redis.Cmdable
}

// NewAuthService 构造函数
func NewAuthService(userRepo repository.UserRepository, jwtManager JwtManager, passHasher PasswordHasher, idGen IDGenerator, client redis.Cmdable) AuthService {
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
		return empty, err
	}

	// 构造指针
	u := &model.User{
		ID:           svc.idGen.NextID(),
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
	}

	// 创建记录
	err = svc.userRepo.Create(ctx, u)
	if err != nil {
		return empty, toErrnoErr(err)
	}

	return userdto.ToBriefDTO(u), nil
}

// Login 登录
func (svc *authService) Login(ctx context.Context, username, pass string) (userdto.BriefDTO, error) {
	var empty userdto.BriefDTO

	// 参数校验
	if username == "" || pass == "" {
		return empty, errno.ErrInvalidParam
	}

	// 获取用户
	user, err := svc.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return empty, toErrnoErr(err)
	}
	if user == nil {
		return empty, errno.ErrInvalidCredential
	}

	// 比较密码
	err = svc.passHasher.Compare(user.PasswordHash, pass)
	if err != nil {
		return empty, err
	}

	return userdto.ToBriefDTO(user), nil
}

// Logout 登出
func (svc *authService) Logout(ctx context.Context, accessToken string) error {
	claim, err := svc.jwtManager.VerifyToken(accessToken)
	if err != nil {
		return errno.ErrLogoutFailed
	}
	ssid := claim.SSid

	// 将 < auth:ssid:xxxxxx, > 放入缓存表明解析出有 ssid 的用户已经被踢下线
	err = svc.client.Set(ctx, ClearTokenPrefix+ssid, "", RefreshTokenExpiration).Err()
	if err != nil {
		return errno.ErrLogoutFailed
	}

	return nil
}

// IssueTokens 签发 Token
func (svc *authService) IssueTokens(ctx context.Context, id int64, role int) (string, string, error) {
	// 参数校验
	if role > 1 || role < 0 {
		role = 0
	}

	// 构造 Jwt 中存放的 Claims
	ssid := uuid.New().String()
	claims := JwtClaim{
		ID:   id,
		Role: role,
		SSid: ssid,
	}

	// 生成 AccessToken
	accessToken, err := svc.jwtManager.GenToken(claims, 0)
	if err != nil {
		return "", "", errno.ErrJwtTokenIssueFailed
	}

	// 生成 RefreshToken
	refreshToken := xid.New().String()

	// 将 < auth:refresh:xxxx.xxxx.xxx, xxxx.xxxx.xxx> 放入缓存
	svc.client.Set(ctx, FreshTokenPrefix+refreshToken, accessToken, RefreshTokenExpiration)

	return accessToken, refreshToken, nil
}
