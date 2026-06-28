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
	"github.com/yzletter/go-postery/backend/conf"
	"github.com/yzletter/go-postery/backend/event"
	"github.com/yzletter/go-postery/backend/grpc/errs"
	"github.com/yzletter/go-postery/backend/grpc/manager"
	"github.com/yzletter/go-postery/backend/micro/auth/model"
	"github.com/yzletter/go-postery/backend/micro/auth/repository"
	"github.com/yzletter/go-postery/backend/ports"
)

type authService struct {
	authRepo   repository.AuthRepository
	jwtManager ports.JwtManager
	passHasher ports.PasswordHasher
	idGen      ports.IDGenerator
	codeClient manager.CodeClient
}

// NewAuthService 构造函数
func NewAuthService(authRepo repository.AuthRepository, jwtManager ports.JwtManager, passHasher ports.PasswordHasher, idGen ports.IDGenerator, codeServiceManager manager.CodeClient) AuthService {
	return &authService{
		authRepo:   authRepo,
		jwtManager: jwtManager,
		passHasher: passHasher,
		idGen:      idGen,
		codeClient: codeServiceManager,
	}
}

// LoginByPassword 手机号码 / 邮箱 + 密码登录
func (svc *authService) LoginByPassword(ctx context.Context, identifier string, password string) (int64, error) {
	// 获取用户登录认证
	identity, err := svc.authRepo.GetAuthIdentityByIdentifier(ctx, identifier)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) { // 邮箱或者手机号没有认证过
			slog.Info("login rejected: identity not found")
			return 0, errs.ErrInvalidArgument
		}
		slog.Error("get auth identity failed", "error", err)
		return 0, errs.ErrInternal
	}

	// 获取用户密码
	passwordHash, err := svc.authRepo.GetPasswordHash(ctx, identity.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			// 不应该发生的错误
			slog.Warn("password hash not found", "uid", identity.UserID, "error", err)
			return 0, errs.ErrInvalidArgument
		}
		slog.Error("get password hash failed", "uid", identity.UserID, "error", err)
		return 0, errs.ErrInternal
	}

	// 比较密码
	if err := svc.passHasher.Compare(passwordHash, password); err != nil {
		if errors.Is(err, ports.ErrInvalidPassword) { // 密码错误, 返回为请求参数错误
			slog.Info("login rejected: invalid password", "uid", identity.UserID)
			return 0, errs.ErrInvalidArgument
		}
		slog.Error("compare password failed", "uid", identity.UserID, "error", err)
		return 0, errs.ErrInternal
	}

	if err := svc.ensureUserCanLogin(ctx, identity.UserID); err != nil {
		return 0, err
	}

	return identity.UserID, nil
}

// LoginByPhone 手机号码 + 验证码进行登录, 未注册的手机号码自动进行注册
func (svc *authService) LoginByPhone(ctx context.Context, phone string, code string) (int64, error) {
	// 校验验证码并消费
	verifyReq := code_grpc.CheckCodeRequest{Biz: int64(conf.CodeBizSMS), Identifier: phone, Code: code}
	if resp, err := svc.codeClient.Verify(ctx, &verifyReq); err != nil {
		// 下游挂了
		slog.Error("verify login code failed", "error", err)
		return 0, errs.ErrInternal
	} else if !resp.Result { // 验证码错误
		slog.Info("login rejected: invalid code")
		return 0, errs.ErrInvalidArgument
	}

	// 获取登录认证
	authType := model.AuthTypeFromBiz(conf.CodeBizSMS) // 认证类型
	authIdentity, err := svc.authRepo.GetAuthIdentity(ctx, authType, phone)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			// 用户不存在, 创建用户（包括用户最小项、用户登录认证、无密码、用户资料、注册扩展功能）
			uid := svc.idGen.NextID()
			verifiedAt := time.Now()
			authType := model.AuthTypeFromBiz(conf.CodeBizSMS)

			nickname := newNickname()
			user := model.User{ID: uid}

			// 登录认证方式
			authIdentity := model.AuthIdentity{
				ID:         svc.idGen.NextID(),
				UserID:     uid,
				AuthType:   authType,
				Identifier: phone,
				IsVerified: 1,
				VerifiedAt: &verifiedAt,
			}

			// 用户资料
			userProfile := model.UserProfile{UserID: uid, Nickname: nickname}

			// Event
			events := make([]*event.OutboxEvent, 0)

			// 注册聊天功能 Event
			value, _ := sonic.MarshalString(event.NewUserEventPayload{ID: uid})
			sessionEvent := event.NewKafkaOutboxEvent(svc.idGen.NextID(), event.KafkaSessionTopic, event.KafkaSessionGroup, value)
			events = append(events, sessionEvent)

			// 初始化用户分数 Event
			value2, _ := sonic.MarshalString(event.UpdateScoreEventPayload{
				ID:    svc.idGen.NextID(),
				Biz:   event.UpdateUserScore, // 更新用户分数
				BizID: uid,
			})
			rankEvent := event.NewKafkaOutboxEvent(svc.idGen.NextID(), event.KafkaTopicRankUpdateScore, event.KafkaRankGroup, value2)
			events = append(events, rankEvent)

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
					slog.Warn("create user conflict", "uid", uid, "error", err)
					return 0, errs.ErrAlreadyExits
				}
				slog.Error("create user failed", "uid", uid, "error", err)
				return 0, errs.ErrInternal
			}

			if err := svc.ensureUserCanLogin(ctx, authIdentity.UserID); err != nil {
				return 0, err
			}

			return authIdentity.UserID, nil
		}
		slog.Error("get auth identity failed", "error", err)
		return 0, errs.ErrInternal
	}

	if err := svc.ensureUserCanLogin(ctx, authIdentity.UserID); err != nil {
		return 0, err
	}

	return authIdentity.UserID, nil
}

// HasPassword 查询密码状态
func (svc *authService) HasPassword(ctx context.Context, id int64) (bool, error) {
	if has, err := svc.authRepo.HasPassword(ctx, id); err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) { // 未设置密码, 不是错误
			return false, nil
		}
		slog.Error("check password status failed", "uid", id, "error", err)
		return false, repository.ErrServerInternal
	} else {
		return has, nil
	}
}

// SetPassword 初始化密码
func (svc *authService) SetPassword(ctx context.Context, uid int64, code string, password string) error {
	// 获取当前用户认证的手机号
	authIdentity, err := svc.authRepo.GetAuthIdentityByAuthType(ctx, uid, model.AuthTypeFromBiz(conf.CodeBizSMS))
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			// 不应该出现的错误
			slog.Info("set password rejected: phone not bound", "uid", uid)
			return errs.ErrUnauthenticated
		}
		slog.Error("get auth identity for set password failed", "uid", uid, "error", err)
		return errs.ErrInternal
	}

	// 校验验证码并消费
	verifyReq := code_grpc.CheckCodeRequest{
		Biz:        int64(conf.CodeBizSMS),
		Identifier: authIdentity.Identifier,
		Code:       code,
	}
	if resp, err := svc.codeClient.Verify(ctx, &verifyReq); err != nil {
		// 下游挂了
		slog.Error("verify set password code failed", "uid", uid, "error", err)
		return errs.ErrInternal
	} else if !resp.Result { // 验证码错误
		slog.Info("set password rejected: invalid code", "uid", uid)
		return errs.ErrInvalidArgument
	}

	// 对密码进行哈希
	passwordHash, err := svc.passHasher.Hash(password)
	if err != nil {
		slog.Error("hash password failed", "uid", uid, "error", err)
		return errs.ErrInternal
	}

	// 初始化密码
	if err := svc.authRepo.SetPassword(ctx, &model.AuthPassword{UserID: uid, PasswordHash: passwordHash}); err != nil {
		if errors.Is(err, repository.ErrUniqueKey) {
			// 不应该出现的错误
			slog.Warn("set password conflict", "uid", uid, "error", err)
			return errs.ErrInternal
		}
		slog.Error("set password failed", "uid", uid, "error", err)
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
			slog.Info("update password rejected: password not found", "uid", uid)
			return errs.ErrNotFound
		}
		slog.Error("get old password hash failed", "uid", uid, "error", err)
		return errs.ErrInternal
	}

	// 判断旧密码是否正确
	if err := svc.passHasher.Compare(oldPasswordHash, oldPassword); err != nil {
		if errors.Is(err, ports.ErrInvalidPassword) {
			// 旧密码错误
			slog.Info("update password rejected: invalid old password", "uid", uid)
			return errs.ErrInvalidArgument
		}
		slog.Error("compare old password failed", "uid", uid, "error", err)
		return errs.ErrInternal
	}

	// 对新密码进行加密
	newPassHash, err := svc.passHasher.Hash(newPassword)
	if err != nil {
		slog.Error("hash new password failed", "uid", uid, "error", err)
		return errs.ErrInternal
	}

	// 改新密码
	if err := svc.authRepo.UpdatePasswordHash(ctx, uid, newPassHash); err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			// 不应该出现的错误
			slog.Warn("update password target not found", "uid", uid)
			return errs.ErrNotFound
		}
		slog.Error("update password hash failed", "uid", uid, "error", err)
		return errs.ErrInternal
	}

	return nil
}

// GetAuthIdentityByUID 获取用户身份认证
func (svc *authService) GetAuthIdentityByUID(ctx context.Context, id int64) (string, string, error) {
	phone, email, err := svc.authRepo.GetAuthIdentityByUID(ctx, id)
	if err != nil {
		slog.Error("get auth identity by uid failed", "uid", id, "error", err)
		return phone, email, errs.ErrInternal
	}
	return phone, email, nil
}

// IssueTokens 签发双 Token
func (svc *authService) IssueTokens(ctx context.Context, uid int64, _ int, userAgent string) (string, string, error) {
	// 获取用户最小信息，保证只给真实存在的用户签发 Token
	user, err := svc.authRepo.GetUser(ctx, uid)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			slog.Info("issue token rejected: user not found", "uid", uid)
			return "", "", errs.ErrUnauthenticated
		}
		slog.Error("get user for issue token failed", "uid", uid, "error", err)
		return "", "", errs.ErrInternal
	}

	// 用户已被禁用或逻辑删除时拒绝签发 Token
	if user.Status != model.UserStatusNormal || user.DeletedAt != nil {
		slog.Info("issue token rejected: user disabled", "uid", uid, "status", user.Status)
		return "", "", errs.ErrUnauthenticated
	}

	// 使用用户表中的真实 Role，不信任调用方传入的 Role
	role := user.Role
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
		slog.Error("generate access token failed", "uid", uid, "error", err)
		return "", "", errs.ErrInternal
	}

	// 生成 RefreshToken
	refreshToken := xid.New().String()

	// 将 < auth:refresh:xxxxxx, uid, ssid, role > 存入缓存
	mp := map[string]any{"user_id": uid, "ssid": ssid, "role": role}
	if err := svc.authRepo.SetInfo(ctx, refreshToken, mp); err != nil {
		slog.Error("set refresh token info failed", "uid", uid, "error", err)
		return accessToken, "", errs.ErrInternal
	}
	return accessToken, refreshToken, nil
}

// ClearTokens 清除双 Token
func (svc *authService) ClearTokens(ctx context.Context, accessToken string, refreshToken string) error {
	// 删除 refreshToken
	if refreshToken != "" {
		if err := svc.authRepo.DelRefreshToken(ctx, refreshToken); err != nil {
			slog.Error("delete refresh token failed", "error", err)
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
		slog.Info("access token rejected", "error", err)
		return &ports.JWTTokenClaims{}, errs.ErrUnauthenticated
	}
	return claim, nil
}

// GetInfoByRefreshToken 根据 RefreshToken 获取用户信息, 用于重新签发双 Token
func (svc *authService) GetInfoByRefreshToken(ctx context.Context, refreshToken string) (int64, int, string, error) {
	uid, role, ssid, err := svc.authRepo.GetInfoByRefreshToken(ctx, refreshToken)
	if err != nil {
		// RefreshToken 不存在、过期、数据损坏都视为认证失败
		if errors.Is(err, repository.ErrRecordNotFound) || errors.Is(err, repository.ErrInvalidToken) {
			slog.Info("refresh token rejected", "error", err)
			return 0, 0, "", errs.ErrUnauthenticated
		}
		slog.Error("get refresh token info failed", "error", err)
		return 0, 0, "", errs.ErrInternal
	}
	return uid, role, ssid, nil
}

// CheckBlackList 根据 SSID 检查黑名单, 检查用户是否被拉黑
func (svc *authService) CheckBlackList(ctx context.Context, ssid string) (bool, error) {
	exist, err := svc.authRepo.CheckBlackList(ctx, ssid)
	if err != nil {
		slog.Error("check token blacklist failed", "error", err)
		return false, errs.ErrInternal
	}
	return exist, nil
}

// 生成默认用户名
func newNickname() string {
	return "用户_" + uuid.NewString()[:8]
}

// 校验用户状态能否进行登录
func (svc *authService) ensureUserCanLogin(ctx context.Context, uid int64) error {
	user, err := svc.authRepo.GetUser(ctx, uid)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			slog.Info("login rejected: user not found", "uid", uid)
			return errs.ErrUnauthenticated
		}
		slog.Error("get user for login failed", "uid", uid, "error", err)
		return errs.ErrInternal
	}

	if user.Status != model.UserStatusNormal || user.DeletedAt != nil {
		slog.Info("login rejected: user disabled", "uid", uid, "status", user.Status)
		return errs.ErrUnauthenticated
	}

	return nil
}
