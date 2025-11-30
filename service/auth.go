package service

import (
	"encoding/json"
	"log/slog"

	"github.com/go-redis/redis"
	"github.com/yzletter/go-postery/dto"
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
func (service *AuthService) GetUserInfoFromJWT(jwtToken string) *dto.UserInformation {
	// 校验 JWT Token
	payload, err := service.JwtService.VerifyToken(jwtToken)
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
			var userInfo dto.UserInformation
			_ = json.Unmarshal(bs, &userInfo)
			slog.Info("AuthService 获得 UserInfo 成功 ... ", "userInfo", userInfo)
			return &userInfo
		}
	}

	// 未获得用户信息
	slog.Info("AuthService 获得 UserInfo 失败 ... ")
	return nil
}
