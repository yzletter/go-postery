package auth

import (
	"encoding/json"
	"log/slog"

	"github.com/go-redis/redis"
	"github.com/yzletter/go-postery/dto"
	"github.com/yzletter/go-postery/service"
)

// AuthHandler 鉴权中间件的 Handler
type AuthHandler struct {
	JwtService  *service.JwtService // 依赖 JWT 相关服务
	RedisClient redis.Cmdable       // 依赖 Redis 数据库
}

// NewAuthHandler 构造函数
func NewAuthHandler(redisClient redis.Cmdable, jwtService *service.JwtService) *AuthHandler {
	return &AuthHandler{
		JwtService:  jwtService,
		RedisClient: redisClient,
	}
}

// GetUserInfoFromJWT 从 JWT Token 中获取 uid
func (authHandler *AuthHandler) GetUserInfoFromJWT(jwtToken string) *dto.UserInformation {
	// 校验 JWT Token
	payload, err := authHandler.JwtService.VerifyToken(jwtToken)
	if err != nil { // JWT Token 校验失败
		slog.Error("AuthHandler 校验 JWT Token 失败 ...", "err", err)
		return nil
	}

	slog.Info("AuthHandler 校验 JWT Token 成功 ...", "payload", payload)

	// 从 payload 的自定义字段中获取用户信息
	for k, v := range payload.UserDefined {
		if k == USERINFO_IN_JWT_PAYLOAD {
			// todo 待优化
			bs, _ := json.Marshal(v)
			var userInfo dto.UserInformation
			_ = json.Unmarshal(bs, &userInfo)
			slog.Info("AuthHandler 获得 UserInfo 成功 ... ", "userInfo", userInfo)
			return &userInfo
		}
	}

	// 未获得用户信息
	slog.Info("AuthHandler 获得 UserInfo 失败 ... ")
	return nil
}
