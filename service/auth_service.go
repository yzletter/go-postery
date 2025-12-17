package service

import (
	"context"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/xid"
	"github.com/yzletter/go-postery/conf"
	"github.com/yzletter/go-postery/model"

	"time"

	"github.com/redis/go-redis/v9"
	userdto "github.com/yzletter/go-postery/dto/user"
	"github.com/yzletter/go-postery/errno"
	"github.com/yzletter/go-postery/repository"
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

// ClearTokens 登出
func (svc *authService) ClearTokens(ctx context.Context, accessToken, refreshToken string) error {
	if accessToken != "" {
		claim, err := svc.VerifyToken(accessToken)
		if err != nil {
			return errno.ErrLogoutFailed
		}
		ssid := claim.SSid

		// 将 < auth:ssid:xxxxxx, > 放入缓存表明解析出有 ssid 的用户已经被踢下线
		err = svc.client.Set(ctx, conf.ClearTokenPrefix+ssid, "", conf.RefreshTokenCookieMaxAgeSecs).Err()
		if err != nil {
			return errno.ErrLogoutFailed
		}
	}

	// 将刷新 Token 删除
	svc.client.Del(ctx, conf.RefreshTokenPrefix+refreshToken)

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
	accessClaims := JWTTokenClaims{
		Uid:       id,
		SSid:      ssid,
		Role:      role,
		UserAgent: agent,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "go-postery",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(conf.AccessTokenExpiration * time.Second)),
		},
	}

	// 生成 AccessToken
	accessToken, err := svc.GenToken(accessClaims)
	if err != nil {
		return "", "", errno.ErrJwtTokenIssueFailed
	}

	// 生成 RefreshTokenCookieMaxAgeSecs
	refreshToken := xid.New().String()

	// 将 < auth:refresh:xxxxxx, ssid > 存入
	mp := map[string]any{
		"user_id": id,
		"ssid":    ssid,
		"role":    role,
	}
	svc.client.HSet(ctx, conf.RefreshTokenPrefix+refreshToken, mp)
	svc.client.Expire(ctx, conf.RefreshTokenPrefix+refreshToken, conf.RefreshTokenCookieMaxAgeSecs)

	return accessToken, refreshToken, nil
}

func (svc *authService) GenToken(claim JWTTokenClaims) (string, error) {
	return svc.jwtManager.GenToken(claim)
}

func (svc *authService) VerifyToken(tokenString string) (*JWTTokenClaims, error) {
	return svc.jwtManager.VerifyToken(tokenString)
}
