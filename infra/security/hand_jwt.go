package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"github.com/yzletter/go-postery/errno"
	"github.com/yzletter/go-postery/service"
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

type HandwrittenJwtManager struct {
	Issuer string
	Secret string        // 用于签名加密的 Secret
	Header JwtHeader     // 默认的 JWT Header
	Leeway time.Duration // 可容忍的时间偏移
}

func NewHandwrittenJwtManager(issuer string, secret string, seconds int) service.JwtManager {
	return &HandwrittenJwtManager{
		Issuer: issuer,
		Secret: secret,
		Header: JwtHeader{ // 默认的 JWT Header
			Algo: "HS256",
			Type: "JWT",
		},
		Leeway: time.Duration(seconds) * time.Second,
	}
}

// GenToken 根据信息生成 JWT Token
func (m *HandwrittenJwtManager) GenToken(claim service.JwtClaim, expiration int64) (string, error) {
	// 参数校验
	if m.Secret == "" {
		return "", errno.ErrJwtInvalidParam
	}

	// 1. header 转成 json, 再用 base64 编码, 得到 JWT 第一部分
	part1, err := marshalBase64Encode(m.Header)
	if err != nil {
		return "", err
	}

	userDefined := map[string]any{
		"id":        claim.ID,   // 用户 ID
		"role":      claim.Role, // 用户角色
		"sessionID": claim.SSid, // 用于废弃该 Token
	}

	// 2. payload 转成 json, 再用 base64 编码, 得到 JWT 第二部分
	payload := JwtPayload{
		ID:          "rj4t49tu49",
		Issue:       m.Issuer,
		Audience:    "王者荣耀",
		Subject:     "购买道具",
		IssueAt:     time.Now().Unix(), // 签发日期为当前时间
		Expiration:  expiration,        // 永不过期
		UserDefined: userDefined,       // 用户自定义字段
	}

	part2, err := marshalBase64Encode(payload)
	if err != nil {
		return "", err
	}

	// 3. 根据 msg 使用 secret 进行加密得到签名 signature
	jwtMsg := part1 + "." + part2                // JWT 信息部分
	jwtSignature := signSha256(jwtMsg, m.Secret) // JWT 签名部分

	return jwtMsg + "." + jwtSignature, nil
}

// VerifyToken 校验 JWT Token, 获得信息
func (m *HandwrittenJwtManager) VerifyToken(token string) (*service.JwtClaim, error) {
	// 参数校验
	if token == "" || m.Secret == "" {
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
	thisSignature := signSha256(jwtMsg, m.Secret)
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
	if payload.IssueAt > 0 && now.Add(m.Leeway).Unix() < payload.IssueAt {
		// 当前时间(加上漂移量) < 签名时间, 签在未来
		return nil, errno.ErrJwtInvalidTime
	}
	if payload.NotBefore > 0 && now.Add(m.Leeway).Unix() < payload.NotBefore {
		// 当前时间(加上漂移量) > 生效时间, 还未生效
		return nil, errno.ErrJwtInvalidTime
	}
	if payload.Expiration > 0 && now.Add(-m.Leeway).Unix() > payload.Expiration {
		// 当前时间(减去漂移量) > 过期时间，已经过期
		return nil, errno.ErrJwtInvalidTime
	}

	slog.Info("verify payload", payload)

	claim := &service.JwtClaim{}
	bs, err := json.Marshal(payload.UserDefined)
	if err != nil {
		return nil, errno.ErrJwtMarshalFailed
	}
	err = json.Unmarshal(bs, claim)
	if err != nil {
		return nil, errno.ErrJwtUnMarshalFailed
	}
	return claim, nil
}

// 对结构体依次进行 json 序列化和 base64 编码
func marshalBase64Encode(v any) (string, error) {
	bs, err := json.Marshal(v)
	if err != nil {
		return "", errno.ErrJwtMarshalFailed
	} else {
		return base64.RawURLEncoding.EncodeToString(bs), nil
	}
}

// 对字符串依次进行 base64 解码和 json 反序列化
func base64DecodeUnmarshal(s string, v any) error {
	bs, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return errno.ErrJwtBase64DecodeFailed
	}
	// 将 bs 反序列化到 v 中
	err = json.Unmarshal(bs, v)
	if err != nil {
		return errno.ErrJwtUnMarshalFailed
	}
	return nil
}

// 用 sha256 哈希算法生成 JWT 签名, 传入 JWT Token 的前两部分和密钥, 返回生成的签名字符串
func signSha256(jwtMsg string, secret string) string {
	hash := hmac.New(sha256.New, []byte(secret))               // 根据 secret 生成 sha256 哈希算法器
	hash.Write([]byte(jwtMsg))                                 // 将 jwtMsg 写入
	return base64.RawURLEncoding.EncodeToString(hash.Sum(nil)) // 对哈希结果进行 base64 编码
}
