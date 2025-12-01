package service

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/go-redis/redis"
	"github.com/rs/xid"
	"github.com/yzletter/go-postery/dto/request"
)

const (
	UID_IN_CTX                = "uid" // uid 在上下文中的 name
	UNAME_IN_CTX              = "uname"
	ACCESS_TOKEN_COOKIE_NAME  = "jwt-access-token"  // AccessToken 在 cookie 中的 name
	REFRESH_TOKEN_COOKIE_NAME = "jwt-refresh-token" // RefreshToken 在 cookie 中的 name
	USERINFO_IN_JWT_PAYLOAD   = "userInfo"
	REFRESH_KEY_PREFIX        = "session_"
)

// AuthService 鉴权中间件的 Service
type AuthService struct {
	JwtService  *JwtService   // 依赖 JWT 相关服务
	RedisClient redis.Cmdable // 依赖 Redis 数据库
}

// NewAuthService 构造函数
func NewAuthService(redisClient redis.Cmdable, jwtService *JwtService) *AuthService {
	return &AuthService{
		JwtService:  jwtService,
		RedisClient: redisClient,
	}
}

// GetUserInfoFromJWT 从 JWT Token 中获取 uid
func (svc *AuthService) GetUserInfoFromJWT(jwtToken string) *request.UserInformation {
	// 校验 JWT Token
	payload, err := svc.JwtService.VerifyToken(jwtToken)
	if err != nil { // JWT Token 校验失败
		slog.Error("AuthService 校验 JWT Token 失败 ...", "err", err)
		return nil
	}

	slog.Info("AuthService 校验 JWT Token 成功 ...", "payload", payload)

	// 从 payload 的自定义字段中获取用户信息
	for k, v := range payload.UserDefined {
		if k == USERINFO_IN_JWT_PAYLOAD {
			// todo 待优化
			bs, _ := json.Marshal(v)
			var userInfo request.UserInformation
			_ = json.Unmarshal(bs, &userInfo)
			slog.Info("AuthService 获得 UserInfo 成功 ... ", "userInfo", userInfo)
			return &userInfo
		}
	}

	// 未获得用户信息
	slog.Info("AuthService 获得 UserInfo 失败 ... ")
	return nil
}

func (svc *AuthService) IssueTokenPairForUser(userInfo request.UserInformation) (string, string, error) {
	// 生成 RefreshToken
	refreshToken := xid.New().String() //	生成一个随机的字符串

	// 生成 AccessToken
	payload := JwtPayload{
		Issue:       "yzletter",
		IssueAt:     time.Now().Unix(),                                 // 签发日期为当前时间
		Expiration:  0,                                                 // 永不过期
		UserDefined: map[string]any{USERINFO_IN_JWT_PAYLOAD: userInfo}, // 用户自定义字段
	}
	accessToken, err := svc.JwtService.GenToken(payload)
	if err != nil {
		return "", "", err
	}

	// < session_refreshToken, accessToken > 放入 redis
	svc.RedisClient.Set(REFRESH_KEY_PREFIX+refreshToken, accessToken, 7*86400*time.Second)

	return refreshToken, accessToken, nil
}
