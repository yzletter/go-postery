package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	"github.com/rs/xid"
	auth_grpc "github.com/yzletter/go-postery/api/proto/auth/v1"
	code_grpc "github.com/yzletter/go-postery/api/proto/code/v1"
	"github.com/yzletter/go-postery/auth/conf"
	"github.com/yzletter/go-postery/auth/model"
	"github.com/yzletter/go-postery/auth/repository"
	"github.com/yzletter/go-postery/auth/service/ports"
	code_conf "github.com/yzletter/go-postery/code/conf"
	code_model "github.com/yzletter/go-postery/code/model"
	"github.com/yzletter/go-postery/errno"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type authService struct {
	authRepo   repository.AuthRepository
	jwtManager ports.JwtManager
	passHasher ports.PasswordHasher
	idGen      ports.IDGenerator

	codeConn *grpc.ClientConn

	auth_grpc.UnimplementedAuthServiceServer
}

// NewAuthService 构造函数
func NewAuthService(authRepo repository.AuthRepository, jwtManager ports.JwtManager, passHasher ports.PasswordHasher, idGen ports.IDGenerator) AuthService {
	codeConn := newCodeGrpcConn()
	return &authService{
		authRepo:                       authRepo,
		jwtManager:                     jwtManager,
		passHasher:                     passHasher,
		idGen:                          idGen,
		codeConn:                       codeConn,
		UnimplementedAuthServiceServer: auth_grpc.UnimplementedAuthServiceServer{},
	}
}

// LoginByPassword 手机号码/邮箱 + 密码登录
func (svc *authService) LoginByPassword(ctx context.Context, req *auth_grpc.LoginByPasswordRequest) (*auth_grpc.UserID, error) {
	var empty = new(auth_grpc.UserID)

	// 获取登录认证
	authIdentity, err := svc.authRepo.GetAuthIdentityByIdentifier(ctx, req.Identifier)
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
	err = svc.passHasher.Compare(passwordHash, req.Password)
	if err != nil {
		if errors.Is(err, ports.ErrInvalidPassword) { // 密码错误, 返回为账号或密码错误
			return empty, errno.ErrInvalidCredential
		}
		return empty, errno.ErrServerInternal
	}

	return &auth_grpc.UserID{UserID: authIdentity.UserID}, nil
}

// LoginByPhone 手机号码 + 验证码进行登录, 未注册的手机号码自动进行注册
func (svc *authService) LoginByPhone(ctx context.Context, req *auth_grpc.LoginByPhoneRequest) (*auth_grpc.UserID, error) {
	var empty = new(auth_grpc.UserID)

	// 判断连接是否可复用
	if !(svc.codeConn.GetState() == connectivity.Ready || svc.codeConn.GetState() == connectivity.Connecting) {
		_ = svc.codeConn.Close() // 关闭旧连接
		svc.codeConn = newCodeGrpcConn()
	}
	codeClient := code_grpc.NewCodeServiceClient(svc.codeConn)

	// 校验验证码并消费
	verifyReq := code_grpc.CheckCodeRequest{
		Biz:        int64(code_model.SMSCode),
		Identifier: req.Phone,
		Code:       req.Code,
	}
	resp, err := codeClient.Verify(ctx, &verifyReq)
	if err != nil {
		slog.Error("Check Code Failed", "biz", code_model.SMSCode)
		return empty, errno.ErrServerInternal
	} else if !resp.Result {
		return empty, errno.ErrPhoneCodeInvalid
	}

	// 获取登录认证
	authType := model.AuthTypeFromBiz(model.SMSCode)
	authIdentity, err := svc.authRepo.GetAuthIdentity(ctx, authType, req.Phone)
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
				Identifier: req.Phone,
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
					return empty, errno.ErrUserDuplicated
				}
				return empty, errno.ErrServerInternal
			}

			return &auth_grpc.UserID{UserID: authIdentity.UserID}, nil
		}
		return empty, errno.ErrServerInternal
	}

	return &auth_grpc.UserID{UserID: authIdentity.UserID}, nil
}

// HasPassword 查询密码状态
func (svc *authService) HasPassword(ctx context.Context, id *auth_grpc.UserID) (*auth_grpc.HasPasswordResponse, error) {
	has, err := svc.authRepo.HasPassword(ctx, id.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return &auth_grpc.HasPasswordResponse{Result: false}, nil
		}
		return &auth_grpc.HasPasswordResponse{Result: false}, repository.ErrServerInternal
	}
	return &auth_grpc.HasPasswordResponse{Result: has}, nil
}

// SetPassword 初始化密码
func (svc *authService) SetPassword(ctx context.Context, req *auth_grpc.SetPasswordRequest) (*auth_grpc.AuthEmptyResponse, error) {
	// 获取当前用户认证的手机号
	authIdentity, err := svc.authRepo.GetAuthIdentityByAuthType(ctx, req.UserID, model.AuthTypeFromBiz(model.SMSCode))
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			slog.Error("Set Pass Without AuthIdentity", "error", err)
			return &auth_grpc.AuthEmptyResponse{}, errno.ErrServerInternal
		}
		return &auth_grpc.AuthEmptyResponse{}, errno.ErrServerInternal
	}

	// 判断连接是否可复用
	if !(svc.codeConn.GetState() == connectivity.Ready || svc.codeConn.GetState() == connectivity.Connecting) {
		_ = svc.codeConn.Close() // 关闭旧连接
		svc.codeConn = newCodeGrpcConn()
	}
	codeClient := code_grpc.NewCodeServiceClient(svc.codeConn)

	// 校验验证码并消费
	verifyReq := code_grpc.CheckCodeRequest{
		Biz:        int64(model.SMSCode),
		Identifier: authIdentity.Identifier,
		Code:       req.Code,
	}
	resp, err := codeClient.Verify(ctx, &verifyReq)
	if err != nil {
		slog.Error("Check Code Failed", "biz", model.SMSCode)
		return &auth_grpc.AuthEmptyResponse{}, errno.ErrServerInternal
	} else if !resp.Result {
		return &auth_grpc.AuthEmptyResponse{}, errno.ErrPhoneCodeInvalid
	}

	// 对密码进行哈希
	passwordHash, err := svc.passHasher.Hash(req.Password)
	if err != nil {
		return &auth_grpc.AuthEmptyResponse{}, errno.ErrServerInternal
	}

	// 初始化密码
	var authPassword = model.AuthPassword{
		UserID:       req.UserID,
		PasswordHash: passwordHash,
	}
	if err := svc.authRepo.SetPassword(ctx, &authPassword); err != nil {
		if errors.Is(err, repository.ErrUniqueKey) {
			slog.Error("Set Password Failed", "error", err)
			return &auth_grpc.AuthEmptyResponse{}, errno.ErrSetPassword
		}
		return &auth_grpc.AuthEmptyResponse{}, errno.ErrServerInternal
	}

	return &auth_grpc.AuthEmptyResponse{}, nil
}

// UpdatePassword 更新密码
func (svc *authService) UpdatePassword(ctx context.Context, req *auth_grpc.UpdatePasswordRequest) (*auth_grpc.AuthEmptyResponse, error) {
	// 获取旧密码
	oldPasswordHash, err := svc.authRepo.GetPasswordHash(ctx, req.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return &auth_grpc.AuthEmptyResponse{}, errno.ErrUserNotFound
		}
		return &auth_grpc.AuthEmptyResponse{}, errno.ErrServerInternal
	}

	// 判断旧密码是否正确
	if err := svc.passHasher.Compare(oldPasswordHash, req.OldPassword); err != nil {
		if errors.Is(err, ports.ErrInvalidPassword) {
			return &auth_grpc.AuthEmptyResponse{}, errno.ErrOldPasswordInvalid
		}
		return &auth_grpc.AuthEmptyResponse{}, errno.ErrServerInternal
	}

	// 对新密码进行加密
	newPassHash, err := svc.passHasher.Hash(req.NewPassword)
	if err != nil {
		return &auth_grpc.AuthEmptyResponse{}, errno.ErrServerInternal
	}

	// 改新密码
	if err := svc.authRepo.UpdatePasswordHash(ctx, req.UserID, newPassHash); err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return &auth_grpc.AuthEmptyResponse{}, errno.ErrUserNotFound
		}
		return &auth_grpc.AuthEmptyResponse{}, errno.ErrServerInternal
	}

	return &auth_grpc.AuthEmptyResponse{}, nil
}

// GetAuthIdentityByUID 获取用户身份认证
func (svc *authService) GetAuthIdentityByUID(ctx context.Context, id *auth_grpc.UserID) (*auth_grpc.AuthIdentity, error) {
	phone, email, err := svc.authRepo.GetAuthIdentityByUID(ctx, id.UserID)
	if err != nil {
		return &auth_grpc.AuthIdentity{Phone: "", Email: ""}, errno.ErrServerInternal
	}
	return &auth_grpc.AuthIdentity{Phone: phone, Email: email}, nil
}

// IssueTokens 签发双 Token
func (svc *authService) IssueTokens(ctx context.Context, req *auth_grpc.IssueTokenRequest) (*auth_grpc.DualTokens, error) {
	// 参数校验
	if req.Role > 1 || req.Role < 0 {
		req.Role = 0
	}

	// AccessToken 的 Claims
	ssid := uuid.New().String()
	expir := time.Now().Add(conf.AccessTokenExpiration * time.Second)
	accessClaims := ports.JWTTokenClaims{
		Uid:       req.UserID,
		SSid:      ssid,
		Role:      int(req.Role),
		UserAgent: req.UserAgent,
		Issuer:    "go-postery",
		ExpiresAt: &expir,
	}

	// 生成 AccessToken
	accessToken, err := svc.jwtManager.GenToken(accessClaims)
	if err != nil {
		return &auth_grpc.DualTokens{AccessToken: "", RefreshToken: ""}, errno.ErrServerInternal
	}

	// 生成 RefreshToken
	refreshToken := xid.New().String()

	// 将 < auth:refresh:xxxxxx, ssid > 存入
	mp := map[string]any{
		"user_id": req.UserID,
		"ssid":    ssid,
		"role":    req.Role,
	}
	err = svc.authRepo.SetInfo(ctx, refreshToken, mp)
	if err != nil {
		return &auth_grpc.DualTokens{AccessToken: "", RefreshToken: ""}, errno.ErrServerInternal

	}
	return &auth_grpc.DualTokens{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

// ClearTokens 清除双 Token
func (svc *authService) ClearTokens(ctx context.Context, tokens *auth_grpc.DualTokens) (*auth_grpc.AuthEmptyResponse, error) {
	// 删除 refreshToken
	if tokens.RefreshToken != "" {
		if err := svc.authRepo.DelRefreshToken(ctx, tokens.RefreshToken); err != nil {
			return &auth_grpc.AuthEmptyResponse{}, errno.ErrLogoutFailed
		}
	}
	// 拉黑 ssid
	if tokens.AccessToken != "" {
		if claim, err := svc.jwtManager.VerifyToken(tokens.AccessToken); err == nil && claim != nil && claim.SSid != "" {
			_ = svc.authRepo.SetBlackList(ctx, claim.SSid) // 拉黑 ssid
		}
		// accessToken 解析失败就跳过，不影响 logout 成功
	}

	return &auth_grpc.AuthEmptyResponse{}, nil
}

// VerifyAccessToken 校验 AccessToken
func (svc *authService) VerifyAccessToken(ctx context.Context, token *auth_grpc.AccessToken) (*auth_grpc.JWTTokenClaims, error) {
	claim, err := svc.jwtManager.VerifyToken(token.AccessToken)
	if err != nil {
		return &auth_grpc.JWTTokenClaims{}, errno.ErrUnauthorized
	}
	return &auth_grpc.JWTTokenClaims{
		UserID:    claim.Uid,
		SSID:      claim.SSid,
		Role:      int32(claim.Role),
		UserAgent: claim.UserAgent,
		Issuer:    claim.Issuer,
		Subject:   claim.Subject,
		Audience:  claim.Audience,
		ExpiresAt: TimeToPB(claim.ExpiresAt),
		NotBefore: TimeToPB(claim.NotBefore),
		IssuedAt:  TimeToPB(claim.IssuedAt),
		ID:        claim.ID,
	}, nil
}

// GetInfoByRefreshToken 根据 RefreshToken 获取用户信息, 用于重新签发双 Token
func (svc *authService) GetInfoByRefreshToken(ctx context.Context, token *auth_grpc.RefreshToken) (*auth_grpc.GetInfoByRefreshTokenResponse, error) {
	uid, role, ssid, err := svc.authRepo.GetInfoByRefreshToken(ctx, token.RefreshToken)
	if err != nil {
		return &auth_grpc.GetInfoByRefreshTokenResponse{UserID: 0, Role: 0, SSID: ""}, errno.ErrServerInternal
	}
	return &auth_grpc.GetInfoByRefreshTokenResponse{UserID: uid, Role: int32(role), SSID: ssid}, nil
}

// CheckBlackList 根据 SSID 检查黑名单, 检查用户是否被拉黑
func (svc *authService) CheckBlackList(ctx context.Context, req *auth_grpc.CheckBlackListRequest) (*auth_grpc.CheckBlackListResponse, error) {
	exist, err := svc.authRepo.CheckBlackList(ctx, req.SSID)
	if err != nil {
		return &auth_grpc.CheckBlackListResponse{Result: false}, errno.ErrServerInternal
	}
	return &auth_grpc.CheckBlackListResponse{Result: exist}, nil
}

// 生成默认用户名
func newNickname() string {
	return "用户_" + uuid.NewString()[:8]
}

func newCodeGrpcConn() *grpc.ClientConn {
	conn, err := grpc.NewClient(
		"localhost:"+code_conf.Port,
		grpc.WithTransportCredentials(insecure.NewCredentials()), // 设置传输安全
	)
	if err != nil {
		return nil
	}

	return conn
}

func TimeToPB(t *time.Time) *timestamppb.Timestamp {
	if t == nil || t.IsZero() {
		return nil
	}
	return timestamppb.New(*t)
}
