package service

import (
	"encoding/json"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/xid"
	"github.com/yzletter/go-postery/dto/user"
	"github.com/yzletter/go-postery/errno"
)

// 可容忍的时间偏移
const defaultLeeway = 5 * time.Second
const (
	UID_IN_CTX                = "uid" // uid 在上下文中的 name
	UNAME_IN_CTX              = "uname"
	ACCESS_TOKEN_COOKIE_NAME  = "jwt-access-token"  // AccessToken 在 cookie 中的 name
	REFRESH_TOKEN_COOKIE_NAME = "jwt-refresh-token" // RefreshToken 在 cookie 中的 name
	USERINFO_IN_JWT_PAYLOAD   = "userInfo"
	REFRESH_KEY_PREFIX        = "session_"
)

type JwtPayload struct {
	ID          string         `json:"jti"` // JWT ID
	Issue       string         `json:"iss"` // 签发者
	Audience    string         `json:"aud"` // 受众
	Subject     string         `json:"sub"` // 主题
	IssueAt     int64          `json:"iat"` // 签发时间（秒）
	NotBefore   int64          `json:"nbf"` // 生效时间（秒）
	Expiration  int64          `json:"exp"` // 过期时间（秒），0=永不过期
	UserDefined map[string]any `json:"ud"`  // 自定义字段
}
type JwtHeader struct {
	Algo string `json:"alg"` // 哈希算法, HS256
	Type string `json:"typ"` // JWT
}

// JwtService 鉴权中间件的 Service
type JwtService struct {
	RedisClient redis.Cmdable // 依赖 Redis 数据库
	Secret      string        // 用于签名加密的 Secret
	Header      JwtHeader
}

// NewJwtService 构造函数
func NewJwtService(redisClient redis.Cmdable, secret string) *JwtService {
	return &JwtService{
		RedisClient: redisClient,
		Secret:      secret,
		// 默认的 JWT Header
		Header: JwtHeader{
			Algo: "HS256",
			Type: "JWT",
		},
	}
}

// GetUserInfoFromJWT 从 JWT Token 中获取 uid
func (svc *JwtService) GetUserInfoFromJWT(jwtToken string) *user.JWTInfo {
	// 校验 JWT Token
	payload, err := svc.VerifyToken(jwtToken)
	if err != nil { // JWT Token 校验失败
		slog.Error("JwtService 校验 JWT Token 失败 ...", "err", err)
		return nil
	}

	slog.Info("JwtService 校验 JWT Token 成功 ...", "payload", payload)

	// 从 payload 的自定义字段中获取用户信息
	for k, v := range payload.UserDefined {
		if k == USERINFO_IN_JWT_PAYLOAD {
			// todo 待优化
			bs, _ := json.Marshal(v)
			var userInfo user.JWTInfo
			_ = json.Unmarshal(bs, &userInfo)
			slog.Info("JwtService 获得 UserInfo 成功 ... ", "userInfo", userInfo)
			return &userInfo
		}
	}

	// 未获得用户信息
	slog.Info("JwtService 获得 UserInfo 失败 ... ")
	return nil
}

// IssueTokenForUser 为 User 签发双 token
func (svc *JwtService) IssueTokenForUser(uid int64, uname string) (string, string, error) {
	userInfo := user.JWTInfo{
		Id:   strconv.Itoa(int(uid)),
		Name: uname,
	}

	// 生成 RefreshToken
	refreshToken := xid.New().String() //	生成一个随机的字符串

	// 生成 AccessToken
	payload := JwtPayload{
		Issue:       "yzletter",
		IssueAt:     time.Now().Unix(),                                 // 签发日期为当前时间
		Expiration:  0,                                                 // 永不过期
		UserDefined: map[string]any{USERINFO_IN_JWT_PAYLOAD: userInfo}, // 用户自定义字段
	}

	accessToken, err := svc.GenToken(payload)
	if err != nil {
		return "", "", err
	}

	// < session_refreshToken, accessToken > 放入 redis
	svc.RedisClient.Set(REFRESH_KEY_PREFIX+refreshToken, accessToken, 7*86400*time.Second)

	return refreshToken, accessToken, nil
}

// 根据 payload 生成 JWT Token
func (svc *JwtService) genToken(payload JwtPayload) (string, error) {
	// 参数校验
	if svc.Secret == "" {
		return "", errno.ErrJwtInvalidParam
	}

	// 1. header 转成 json, 再用 base64 编码, 得到 JWT 第一部分
	part1, err := marshalBase64Encode(svc.Header)
	if err != nil {
		return "", err
	}

	// 2. payload 转成 json, 再用 base64 编码, 得到 JWT 第二部分
	part2, err := marshalBase64Encode(payload)
	if err != nil {
		return "", err
	}

	// 3. 根据 msg 使用 secret 进行加密得到签名 signature
	jwtMsg := part1 + "." + part2                  // JWT 信息部分
	jwtSignature := signSha256(jwtMsg, svc.Secret) // JWT 签名部分

	return jwtMsg + "." + jwtSignature, nil
}

// 校验 JWT Token, 获得 payload
func (svc *JwtService) verifyToken(token string) (*JwtPayload, error) {
	// 参数校验
	if token == "" || svc.Secret == "" {
		return nil, errno.ErrJwtInvalidParam
	}
	parts := strings.SplitN(token, ".", 3)
	if len(parts) != 3 {
		// 传入的 JWT 格式有误
		return nil, errno.ErrJwtInvalidParam
	}

	// 获得 msg 和 signature 部分
	jwtMsg := parts[0] + "." + parts[1]
	jwtSignature := parts[2]

	// 1. 签名校验
	// 对 jwtMsg 加密得到 thisSignature 判断与 jwtSignature 是否相同
	thisSignature := signSha256(jwtMsg, svc.Secret)
	if thisSignature != jwtSignature {
		// 签名校验失败
		return nil, errno.ErrJwtInvalidParam
	}

	// 2. 反解出 header 和 payload
	var (
		header  JwtHeader
		payload JwtPayload
	)
	err := base64DecodeUnmarshal(parts[0], &header)
	if err != nil {
		return nil, err
	}
	err = base64DecodeUnmarshal(parts[1], &payload)
	if err != nil {
		return nil, err
	}

	// 3. 时间校验
	now := time.Now()
	if payload.IssueAt > 0 && now.Add(defaultLeeway).Unix() < payload.IssueAt {
		// 当前时间(加上漂移量) < 签名时间, 签在未来
		return nil, errno.ErrJwtInvalidTime
	}
	if payload.NotBefore > 0 && now.Add(defaultLeeway).Unix() < payload.NotBefore {
		// 当前时间(加上漂移量) > 生效时间, 还未生效
		return nil, errno.ErrJwtInvalidTime
	}
	if payload.Expiration > 0 && now.Add(-defaultLeeway).Unix() > payload.Expiration {
		// 当前时间(减去漂移量) > 过期时间，已经过期
		return nil, errno.ErrJwtInvalidTime
	}

	slog.Info("verify payload", payload)
	return &payload, nil
}
