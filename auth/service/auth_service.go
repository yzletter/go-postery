package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	"github.com/rs/xid"
	code_grpc "github.com/yzletter/go-postery/api/proto/code/v1"
	"github.com/yzletter/go-postery/auth/conf"
	"github.com/yzletter/go-postery/auth/errs"
	"github.com/yzletter/go-postery/auth/grpc/client"
	"github.com/yzletter/go-postery/auth/model"
	"github.com/yzletter/go-postery/auth/repository"
	"github.com/yzletter/go-postery/auth/service/ports"
)

type authService struct {
	authRepo   repository.AuthRepository
	jwtManager ports.JwtManager
	passHasher ports.PasswordHasher
	idGen      ports.IDGenerator

	codeClient client.CodeClient
}

// NewAuthService 构造函数
func NewAuthService(authRepo repository.AuthRepository, jwtManager ports.JwtManager, passHasher ports.PasswordHasher, idGen ports.IDGenerator, codeClient client.CodeClient) AuthService {
	return &authService{
		authRepo:   authRepo,
		jwtManager: jwtManager,
		passHasher: passHasher,
		idGen:      idGen,
		codeClient: codeClient,
	}
}

// LoginByPassword 手机号码/邮箱 + 密码登录
func (svc *authService) LoginByPassword(ctx context.Context, identifier string, password string) (int64, error) {
	// 获取登录认证
	authIdentity, err := svc.authRepo.GetAuthIdentityByIdentifier(ctx, identifier)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) { // 邮箱或者手机号没有认证过
			slog.Error("Invalid Identifier")
			return 0, errs.ErrInvalidArgument
		}
		slog.Error("Server Internal Error", "error", err)
		return 0, errs.ErrInternal
	}

	// 获取密码
	passwordHash, err := svc.authRepo.GetPasswordHash(ctx, authIdentity.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			// 不应该发生的错误
			slog.Error("Get Password Failed", "error", err)
			return 0, errs.ErrInvalidArgument
		}
		slog.Error("Server Internal Error", "error", err)
		return 0, errs.ErrInternal
	}

	// 比较密码
	if err := svc.passHasher.Compare(passwordHash, password); err != nil {
		if errors.Is(err, ports.ErrInvalidPassword) { // 密码错误, 返回为请求参数错误
			slog.Error("Invalid Password")
			return 0, errs.ErrInvalidArgument
		}
		slog.Error("Server Internal Error", "error", err)
		return 0, errs.ErrInternal
	}

	return authIdentity.UserID, nil
}

// LoginByPhone 手机号码 + 验证码进行登录, 未注册的手机号码自动进行注册
func (svc *authService) LoginByPhone(ctx context.Context, phone string, code string) (int64, error) {
	// 校验验证码并消费
	verifyReq := code_grpc.CheckCodeRequest{
		Biz:        int64(model.SMSCode),
		Identifier: phone,
		Code:       code,
	}
	resp, err := svc.codeClient.Verify(ctx, &verifyReq)
	if err != nil {
		// 下游挂了
		slog.Error("Code Service Unavailable", "error", err)
		return 0, errs.ErrInternal
	} else if !resp.Result { // 验证码错误
		slog.Error("Invalid Code")
		return 0, errs.ErrInvalidArgument
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
			user := model.User{ID: uid}
			authIdentity := model.AuthIdentity{ // 登录认证方式
				ID:         svc.idGen.NextID(),
				UserID:     uid,
				AuthType:   authType,
				Identifier: phone,
				IsVerified: 1,
				VerifiedAt: &verifiedAt,
			}
			userProfile := model.UserProfile{UserID: uid, Nickname: nickname}
			events := make([]*model.Event, 0)

			// 注册聊天功能 Event
			registerSessionPayload, _ := sonic.Marshal(model.RegisterSessionEventPayload{UserID: uid})
			registerSessionEvent := model.Event{
				ID:           svc.idGen.NextID(),
				Topic:        "session",
				MessageKey:   "register_session",
				MessageValue: string(registerSessionPayload),
			}
			events = append(events, &registerSessionEvent)

			// 初始化用户推荐分数 Event
			initUserScorePayload, _ := sonic.Marshal(model.InitUserScoreEventPayload{UserID: uid})
			initUserScoreEvent := model.Event{
				ID:           svc.idGen.NextID(),
				Topic:        "follow",
				MessageKey:   "init_user_score",
				MessageValue: string(initUserScorePayload),
			}
			events = append(events, &initUserScoreEvent)

			// 聚合信息
			authAggregate := model.AuthAggregate{
				User:         &user,
				UserProfile:  &userProfile,
				AuthPassword: nil, // 无密码
				AuthIdentity: &authIdentity,
				Events:       events,
			}
			if err := svc.authRepo.CreateUser(ctx, &authAggregate); err != nil {
				if errors.Is(err, repository.ErrUniqueKey) {
					// 不应该出现的错误
					slog.Error("Create User Failed", "error", err)
					return 0, errs.ErrAlreadyExits
				}
				slog.Error("Server Internal Error", "error", err)
				return 0, errs.ErrInternal
			}

			return authIdentity.UserID, nil
		}
		slog.Error("Server Internal Error", "error", err)
		return 0, errs.ErrInternal
	}
	return authIdentity.UserID, nil
}

// HasPassword 查询密码状态
func (svc *authService) HasPassword(ctx context.Context, id int64) (bool, error) {
	if has, err := svc.authRepo.HasPassword(ctx, id); err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) { // 未设置密码, 不是错误
			return false, nil
		}
		slog.Error("Server Internal Error", "error", err)
		return false, repository.ErrServerInternal
	} else {
		return has, nil
	}
}

// SetPassword 初始化密码
func (svc *authService) SetPassword(ctx context.Context, uid int64, code string, password string) error {
	// 获取当前用户认证的手机号
	authIdentity, err := svc.authRepo.GetAuthIdentityByAuthType(ctx, uid, model.AuthTypeFromBiz(model.SMSCode))
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			// 不应该出现的错误
			slog.Error("Set Password Without AuthIdentity", "error", err)
			return errs.ErrUnauthenticated
		}
		slog.Error("Server Internal Error", "error", err)
		return errs.ErrInternal
	}

	// 校验验证码并消费
	verifyReq := code_grpc.CheckCodeRequest{
		Biz:        int64(model.SMSCode),
		Identifier: authIdentity.Identifier,
		Code:       code,
	}
	resp, err := svc.codeClient.Verify(ctx, &verifyReq)
	if err != nil {
		// 下游挂了
		slog.Error("Code Service Unavailable", "error", err)
		return errs.ErrInternal
	} else if !resp.Result { // 验证码错误
		slog.Error("Invalid Code")
		return errs.ErrInvalidArgument
	}

	// 对密码进行哈希
	passwordHash, err := svc.passHasher.Hash(password)
	if err != nil {
		slog.Error("Server Internal Error", "error", err)
		return errs.ErrInternal
	}

	// 初始化密码
	var authPassword = model.AuthPassword{
		UserID:       uid,
		PasswordHash: passwordHash,
	}
	if err := svc.authRepo.SetPassword(ctx, &authPassword); err != nil {
		if errors.Is(err, repository.ErrUniqueKey) {
			// 不应该出现的错误
			slog.Error("Set Password Failed", "error", err)
			return errs.ErrInternal
		}
		slog.Error("Server Internal Error", "error", err)
		return errs.ErrInternal
	}

	return nil
}

// UpdatePassword 更新密码
func (svc *authService) UpdatePassword(ctx context.Context, uid int64, oldPassword string, newPassword string) error {
	// 获取旧密码
	oldPasswordHash, err := svc.authRepo.GetPasswordHash(ctx, uid)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			slog.Error("User Not Found")
			return errs.ErrNotFound
		}
		slog.Error("Server Internal Error", "error", err)
		return errs.ErrInternal
	}

	// 判断旧密码是否正确
	if err := svc.passHasher.Compare(oldPasswordHash, oldPassword); err != nil {
		if errors.Is(err, ports.ErrInvalidPassword) {
			// 旧密码错误
			slog.Error("Invalid Old Password")
			return errs.ErrInvalidArgument
		}
		slog.Error("Server Internal Error", "error", err)
		return errs.ErrInternal
	}

	// 对新密码进行加密
	newPassHash, err := svc.passHasher.Hash(newPassword)
	if err != nil {
		slog.Error("Server Internal Error", "error", err)
		return errs.ErrInternal
	}

	// 改新密码
	if err := svc.authRepo.UpdatePasswordHash(ctx, uid, newPassHash); err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			// 不应该出现的错误
			slog.Error("User Not Found")
			return errs.ErrNotFound
		}
		slog.Error("Server Internal Error", "error", err)
		return errs.ErrInternal
	}

	return nil
}

// GetAuthIdentityByUID 获取用户身份认证
func (svc *authService) GetAuthIdentityByUID(ctx context.Context, id int64) (string, string, error) {
	phone, email, err := svc.authRepo.GetAuthIdentityByUID(ctx, id)
	if err != nil {
		slog.Error("Server Internal Error", "error", err)
		return phone, email, errs.ErrInternal
	}
	return phone, email, nil
}

// IssueTokens 签发双 Token
func (svc *authService) IssueTokens(ctx context.Context, uid int64, role int, userAgent string) (string, string, error) {
	// 参数校验
	if role > 1 || role < 0 {
		role = 0
	}

	// AccessToken 的 Claims
	ssid := uuid.New().String()
	expir := time.Now().Add(conf.AccessTokenExpiration * time.Second)
	accessClaims := ports.JWTTokenClaims{
		Uid:       uid,
		SSid:      ssid,
		Role:      role,
		UserAgent: userAgent,
		Issuer:    "go-postery",
		ExpiresAt: &expir,
	}

	// 生成 AccessToken
	accessToken, err := svc.jwtManager.GenToken(accessClaims)
	if err != nil {
		slog.Error("Server Internal Error", "error", err)
		return "", "", errs.ErrInternal
	}

	// 生成 RefreshToken
	refreshToken := xid.New().String()

	// 将 < auth:refresh:xxxxxx, ssid > 存入
	mp := map[string]any{
		"user_id": uid,
		"ssid":    ssid,
		"role":    role,
	}
	err = svc.authRepo.SetInfo(ctx, refreshToken, mp)
	if err != nil {
		slog.Error("Server Internal Error", "error", err)
		return "", "", errs.ErrInternal

	}
	return accessToken, refreshToken, nil
}

// ClearTokens 清除双 Token
func (svc *authService) ClearTokens(ctx context.Context, accessToken string, refreshToken string) error {
	// 删除 refreshToken
	if refreshToken != "" {
		if err := svc.authRepo.DelRefreshToken(ctx, refreshToken); err != nil {
			slog.Error("Server Internal Error", "error", err)
			return errs.ErrInternal
		}
	}

	// 拉黑 ssid
	if accessToken != "" {
		if claim, err := svc.jwtManager.VerifyToken(accessToken); err == nil && claim != nil && claim.SSid != "" {
			_ = svc.authRepo.SetBlackList(ctx, claim.SSid) // 拉黑 ssid
		}
		// accessToken 解析失败就跳过，不影响 logout 成功
	}

	return nil
}

// VerifyAccessToken 校验 AccessToken
func (svc *authService) VerifyAccessToken(ctx context.Context, accessToken string) (*ports.JWTTokenClaims, error) {
	claim, err := svc.jwtManager.VerifyToken(accessToken)
	if err != nil { // AccessToken 校验失败
		slog.Error("Invalid Access Token", "error", err)
		return &ports.JWTTokenClaims{}, errs.ErrUnauthenticated
	}
	return claim, nil
}

// GetInfoByRefreshToken 根据 RefreshToken 获取用户信息, 用于重新签发双 Token
func (svc *authService) GetInfoByRefreshToken(ctx context.Context, refreshToken string) (int64, int, string, error) {
	uid, role, ssid, err := svc.authRepo.GetInfoByRefreshToken(ctx, refreshToken)
	if err != nil {
		slog.Error("Server Internal Error", "error", err)
		return 0, 0, "", errs.ErrInternal
	}
	return uid, role, ssid, nil
}

// CheckBlackList 根据 SSID 检查黑名单, 检查用户是否被拉黑
func (svc *authService) CheckBlackList(ctx context.Context, ssid string) (bool, error) {
	exist, err := svc.authRepo.CheckBlackList(ctx, ssid)
	if err != nil {
		slog.Error("Server Internal Error", "error", err)
		return false, errs.ErrInternal
	}
	return exist, nil
}

// 生成默认用户名
func newNickname() string {
	return "用户_" + uuid.NewString()[:8]
}
