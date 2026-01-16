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

	userdto "github.com/yzletter/go-postery/dto/user"
	"github.com/yzletter/go-postery/errno"
	"github.com/yzletter/go-postery/repository"
)

type authService struct {
	codeSvc      CodeService
	authRepo     repository.AuthRepository
	userRepo     repository.UserRepository
	jwtManager   ports.JwtManager
	emailManager ports.EmailManager
	passHasher   ports.PasswordHasher
	idGen        ports.IDGenerator
}

// NewAuthService 构造函数
func NewAuthService(codeSvc CodeService, authRepo repository.AuthRepository, userRepo repository.UserRepository, jwtManager ports.JwtManager, emailManager ports.EmailManager, passHasher ports.PasswordHasher, idGen ports.IDGenerator) AuthService {
	return &authService{
		codeSvc:      codeSvc,
		authRepo:     authRepo,
		userRepo:     userRepo,
		jwtManager:   jwtManager,
		emailManager: emailManager,
		passHasher:   passHasher,
		idGen:        idGen,
	}
}

// Register 手机号码/邮箱 + 验证码 + 密码注册
func (svc *authService) Register(ctx context.Context, biz model.CodeBiz, identifier, code, password string) (userdto.BriefDTO, error) {
	var empty userdto.BriefDTO

	// 校验验证码并消费
	ok, err := svc.codeSvc.CheckCode(ctx, biz, identifier, code)
	if err != nil {
		slog.Error("Check Code Failed", "biz", biz)
		return empty, errno.ErrServerInternal
	} else if !ok {
		return empty, errno.ErrInvalidCode
	}

	// 创建用户（包括用户最小项、用户登录认证、用户密码、用户资料、todo注册扩展功能）
	uid := svc.idGen.NextID()
	verifiedAt := time.Now()
	passwordHash, err := svc.passHasher.Hash(password) // 对密码进行加密
	if err != nil {
		slog.Error("PasswordHasher Hash Failed", "error", err)
		return empty, errno.ErrServerInternal
	}
	authType := model.AuthTypeFromBiz(biz)
	authIdentity := model.AuthIdentity{ // 登录认证方式
		ID:         svc.idGen.NextID(),
		UserID:     uid,
		AuthType:   authType,
		Identifier: identifier,
		IsVerified: 1,
		VerifiedAt: &verifiedAt,
	}
	if err := svc.authRepo.CreateUser(ctx, &authIdentity, &passwordHash); err != nil {
		if errors.Is(err, repository.ErrUniqueKey) {
			return empty, errno.ErrUserDuplicated
		}
		return empty, errno.ErrServerInternal
	}

	// 查找个人资料
	userProfile, err := svc.userRepo.GetProfileByID(ctx, uid)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return empty, errno.ErrUserNotFound
		}
		return empty, errno.ErrServerInternal
	}

	return userdto.ToBriefDTO(userProfile), nil
}

// LoginByPassword 手机号码/邮箱 + 密码登录
func (svc *authService) LoginByPassword(ctx context.Context, biz model.CodeBiz, identifier, password string) (userdto.BriefDTO, error) {
	var empty userdto.BriefDTO

	// 获取登录认证
	authType := model.AuthTypeFromBiz(biz)
	authIdentity, err := svc.authRepo.GetAuthIdentity(ctx, authType, identifier)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return empty, errno.ErrUserNotFound
		}
		return empty, errno.ErrServerInternal
	}

	uid := authIdentity.UserID // 得到用户 ID

	// 获取密码
	passwordHash, err := svc.authRepo.GetPasswordHash(ctx, uid)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return empty, errno.ErrUserNotFound
		}
		return empty, errno.ErrServerInternal
	}

	// 比较密码
	err = svc.passHasher.Compare(passwordHash, password)
	if err != nil {
		if errors.Is(err, ports.ErrInvalidPassword) { // 密码错误, 返回为账号或密码错误
			return empty, errno.ErrInvalidCredential
		}
		return empty, errno.ErrServerInternal
	}

	// 查找个人资料
	userProfile, err := svc.userRepo.GetProfileByID(ctx, uid)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return empty, errno.ErrUserNotFound
		}
		return empty, errno.ErrServerInternal
	}

	return userdto.ToBriefDTO(userProfile), nil
}

// LoginByEmail 邮箱 + 验证码进行登录
func (svc *authService) LoginByEmail(ctx context.Context, email, code string) (userdto.BriefDTO, error) {
	var empty userdto.BriefDTO

	// 校验验证码并消费
	ok, err := svc.codeSvc.CheckCode(ctx, model.EmailCode, email, code)
	if err != nil {
		slog.Error("Check Code Failed", "biz", model.EmailCode)
		return empty, errno.ErrServerInternal
	} else if !ok {
		return empty, errno.ErrEmailCodeInvalid
	}

	// 获取登录认证
	authType := model.AuthTypeFromBiz(model.EmailCode)
	authIdentity, err := svc.authRepo.GetAuthIdentity(ctx, authType, email)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return empty, errno.ErrUserNotFound
		}
		return empty, errno.ErrServerInternal
	}

	// 查找个人资料
	uid := authIdentity.UserID
	userProfile, err := svc.userRepo.GetProfileByID(ctx, uid)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return empty, errno.ErrUserNotFound
		}
		return empty, errno.ErrServerInternal
	}

	return userdto.ToBriefDTO(userProfile), nil
}

// LoginByPhone 手机号码 + 验证码进行登录, 未注册的手机号码自动进行注册
func (svc *authService) LoginByPhone(ctx context.Context, phone, code string) (userdto.BriefDTO, error) {
	var empty userdto.BriefDTO

	// 校验验证码并消费
	ok, err := svc.codeSvc.CheckCode(ctx, model.SMSCode, phone, code)
	if err != nil {
		slog.Error("Check Code Failed", "biz", model.SMSCode)
		return empty, errno.ErrServerInternal
	} else if !ok {
		return empty, errno.ErrPhoneCodeInvalid
	}

	// 获取登录认证
	authType := model.AuthTypeFromBiz(model.SMSCode)
	authIdentity, err := svc.authRepo.GetAuthIdentity(ctx, authType, phone)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			// 用户不存在, 创建用户（包括用户最小项、用户登录认证、无密码、用户资料、注册扩展功能）
			uid := svc.idGen.NextID()
			verifiedAt := time.Now()
			authIdentity := model.AuthIdentity{ // 登录认证方式
				ID:         svc.idGen.NextID(),
				UserID:     uid,
				AuthType:   authType,
				Identifier: phone,
				IsVerified: 1,
				VerifiedAt: &verifiedAt,
			}

			if err := svc.authRepo.CreateUser(ctx, &authIdentity, nil); err != nil {
				if errors.Is(err, repository.ErrUniqueKey) {
					return empty, errno.ErrUserDuplicated
				}
				return empty, errno.ErrServerInternal
			}

			// 查找个人资料
			userProfile, err := svc.userRepo.GetProfileByID(ctx, uid)
			if err != nil {
				if errors.Is(err, repository.ErrRecordNotFound) {
					return empty, errno.ErrUserNotFound
				}
				return empty, errno.ErrServerInternal
			}

			return userdto.ToBriefDTO(userProfile), nil
		}
		return empty, errno.ErrServerInternal
	}

	// 查找个人资料
	uid := authIdentity.UserID
	userProfile, err := svc.userRepo.GetProfileByID(ctx, uid)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return empty, errno.ErrUserNotFound
		}
		return empty, errno.ErrServerInternal
	}

	return userdto.ToBriefDTO(userProfile), nil
}

// IssueTokens 签发双 Token
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

	// 生成 RefreshToken
	refreshToken := xid.New().String()

	// 将 < auth:refresh:xxxxxx, ssid > 存入
	mp := map[string]any{
		"user_id": id,
		"ssid":    ssid,
		"role":    role,
	}
	err = svc.authRepo.SetInfo(ctx, refreshToken, mp)
	if err != nil {
		return "", "", errno.ErrServerInternal
	}

	return accessToken, refreshToken, nil
}

// ClearTokens 清除双 Token
func (svc *authService) ClearTokens(ctx context.Context, accessToken, refreshToken string) error {
	// 删除 refreshToken
	if refreshToken != "" {
		if err := svc.authRepo.DelRefreshToken(ctx, refreshToken); err != nil {
			return errno.ErrLogoutFailed
		}
	}

	// 拉黑 ssid
	if accessToken != "" {
		if claim, err := svc.VerifyAccessToken(accessToken); err == nil && claim != nil && claim.SSid != "" {
			_ = svc.authRepo.SetBlackList(ctx, claim.SSid) // 拉黑 ssid
		}
		// accessToken 解析失败就跳过，不影响 logout 成功
	}

	return nil
}

// VerifyAccessToken 校验 AccessToken
func (svc *authService) VerifyAccessToken(tokenString string) (*ports.JWTTokenClaims, error) {
	claim, err := svc.jwtManager.VerifyToken(tokenString)
	if err != nil {
		return nil, errno.ErrUnauthorized
	}
	return claim, nil
}

// GetInfoByRefreshToken 根据 RefreshToken 获取用户信息, 用于重新签发双 Token
func (svc *authService) GetInfoByRefreshToken(ctx context.Context, refreshToken string) (int64, int, string, error) {
	uid, role, ssid, err := svc.authRepo.GetInfoByRefreshToken(ctx, refreshToken)
	if err != nil {
		return 0, 0, "", errno.ErrServerInternal
	}
	return uid, role, ssid, nil
}

// CheckBlackList 根据 SSID 检查黑名单, 检查用户是否被拉黑
func (svc *authService) CheckBlackList(ctx context.Context, ssid string) (bool, error) {
	exist, err := svc.authRepo.CheckBlackList(ctx, ssid)
	if err != nil {
		return false, errno.ErrServerInternal
	}
	return exist, nil
}
