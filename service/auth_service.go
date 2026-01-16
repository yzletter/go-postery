package service

import (
	"context"
	"errors"
	"log/slog"

	"time"

	"github.com/google/uuid"
	"github.com/rs/xid"
	"github.com/yzletter/go-postery/conf"
	"github.com/yzletter/go-postery/model"
	"github.com/yzletter/go-postery/service/ports"

	userdto "github.com/yzletter/go-postery/dto/user"
	"github.com/yzletter/go-postery/errno"
	"github.com/yzletter/go-postery/repository"
)

type authService struct {
	codeSvc    CodeService
	authRepo   repository.AuthRepository
	userRepo   repository.UserRepository
	jwtManager ports.JwtManager
	passHasher ports.PasswordHasher
	idGen      ports.IDGenerator
}

// NewAuthService 构造函数
func NewAuthService(codeSvc CodeService, authRepo repository.AuthRepository, userRepo repository.UserRepository, jwtManager ports.JwtManager, passHasher ports.PasswordHasher, idGen ports.IDGenerator) AuthService {
	return &authService{
		codeSvc:    codeSvc,
		authRepo:   authRepo,
		userRepo:   userRepo,
		jwtManager: jwtManager,
		passHasher: passHasher,
		idGen:      idGen,
	}
}

// LoginByPassword 手机号码/邮箱 + 密码登录
func (svc *authService) LoginByPassword(ctx context.Context, identifier, password string) (userdto.BriefDTO, error) {
	var empty userdto.BriefDTO

	// 获取登录认证
	authIdentity, err := svc.authRepo.GetAuthIdentityByIdentifier(ctx, identifier)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return empty, errno.ErrUserNotFound
		}
		return empty, errno.ErrServerInternal
	}

	// 得到用户 ID
	uid := authIdentity.UserID

	// 获取密码
	passwordHash, err := svc.authRepo.GetPasswordHash(ctx, uid)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return empty, errno.ErrInvalidCredential
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
			authType := model.AuthTypeFromBiz(model.SMSCode)

			nickname := newNickname()
			user := model.User{
				ID: uid,
			}
			authIdentity := model.AuthIdentity{ // 登录认证方式
				ID:         svc.idGen.NextID(),
				UserID:     uid,
				AuthType:   authType,
				Identifier: phone,
				IsVerified: 1,
				VerifiedAt: &verifiedAt,
			}
			userProfile := model.UserProfile{UserID: uid, Nickname: nickname}

			// 聚合信息
			authAggregate := model.AuthAggregate{
				User:         &user,
				UserProfile:  &userProfile,
				AuthPassword: nil, // 无密码
				AuthIdentity: &authIdentity,
			}
			if err := svc.authRepo.CreateUser(ctx, &authAggregate); err != nil {
				if errors.Is(err, repository.ErrUniqueKey) {
					return empty, errno.ErrUserDuplicated
				}
				return empty, errno.ErrServerInternal
			}

			return userdto.ToBriefDTO(&userProfile), nil
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

// HasPassword 查询密码状态
func (svc *authService) HasPassword(ctx context.Context, uid int64) (bool, error) {
	has, err := svc.authRepo.HasPassword(ctx, uid)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return false, nil
		}
		return false, repository.ErrServerInternal
	}
	return has, nil
}

// SetPassword 初始化密码
func (svc *authService) SetPassword(ctx context.Context, uid int64, code, newPass string) error {
	// 获取当前用户认证的手机号
	authIdentity, err := svc.authRepo.GetAuthIdentityByAuthType(ctx, uid, model.AuthTypeFromBiz(model.SMSCode))
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			slog.Error("Set Pass Without AuthIdentity", "error", err)
			return errno.ErrServerInternal
		}
		return errno.ErrServerInternal
	}

	// 校验验证码并消费
	ok, err := svc.codeSvc.CheckCode(ctx, model.SMSCode, authIdentity.Identifier, code)
	if err != nil {
		slog.Error("Check Code Failed", "biz", model.SMSCode)
		return errno.ErrServerInternal
	} else if !ok {
		return errno.ErrPhoneCodeInvalid
	}

	// 对密码进行哈希
	passwordHash, err := svc.passHasher.Hash(newPass)
	if err != nil {
		return errno.ErrServerInternal
	}

	// 初始化密码
	var authPassword = model.AuthPassword{
		UserID:       uid,
		PasswordHash: passwordHash,
	}
	if err := svc.authRepo.SetPassword(ctx, &authPassword); err != nil {
		if errors.Is(err, repository.ErrUniqueKey) {
			slog.Error("Set Password Failed", "error", err)
			return errno.ErrSetPassword
		}
		return errno.ErrServerInternal
	}

	return nil
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

// 生成默认用户名
func newNickname() string {
	return "用户_" + uuid.NewString()[:8]
}
