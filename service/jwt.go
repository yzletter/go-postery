package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"
)

type JwtService struct {
	Secret string // 用于签名加密的 Secret
	Header JwtHeader
}

// NewJwtService 构造函数
func NewJwtService(secret string) *JwtService {
	return &JwtService{
		Secret: secret,
		// 默认的 JWT Header
		Header: JwtHeader{
			Algo: "HS256",
			Type: "JWT",
		},
	}
}

// GenToken 根据 payload 生成 JWT Token
func (svc *JwtService) GenToken(payload JwtPayload) (string, error) {
	// 参数校验
	if svc.Secret == "" {
		return "", ErrJwtInvalidParam
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

// VerifyToken 校验 JWT Token, 获得 payload
func (svc *JwtService) VerifyToken(token string) (*JwtPayload, error) {
	// 参数校验
	if token == "" || svc.Secret == "" {
		return nil, ErrJwtInvalidParam
	}
	parts := strings.SplitN(token, ".", 3)
	if len(parts) != 3 {
		// 传入的 JWT 格式有误
		return nil, ErrJwtInvalidParam
	}

	// 获得 msg 和 signature 部分
	jwtMsg := parts[0] + "." + parts[1]
	jwtSignature := parts[2]

	// 1. 签名校验
	// 对 jwtMsg 加密得到 thisSignature 判断与 jwtSignature 是否相同
	thisSignature := signSha256(jwtMsg, svc.Secret)
	if thisSignature != jwtSignature {
		// 签名校验失败
		return nil, ErrJwtInvalidParam
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
		return nil, ErrJwtInvalidTime
	}
	if payload.NotBefore > 0 && now.Add(defaultLeeway).Unix() < payload.NotBefore {
		// 当前时间(加上漂移量) > 生效时间, 还未生效
		return nil, ErrJwtInvalidTime
	}
	if payload.Expiration > 0 && now.Add(-defaultLeeway).Unix() > payload.Expiration {
		// 当前时间(减去漂移量) > 过期时间，已经过期
		return nil, ErrJwtInvalidTime
	}

	return &payload, nil
}

// 对结构体依次进行 json 序列化和 base64 编码
func marshalBase64Encode(v any) (string, error) {
	bs, err := json.Marshal(v)
	if err != nil {
		return "", ErrJwtMarshalFailed
	} else {
		return base64.RawURLEncoding.EncodeToString(bs), nil
	}
}

// 对字符串依次进行 base64 解码和 json 反序列化
func base64DecodeUnmarshal(s string, v any) error {
	bs, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return ErrJwtBase64DecodeFailed
	}
	// 将 bs 反序列化到 v 中
	err = json.Unmarshal(bs, v)
	if err != nil {
		return ErrJwtUnMarshalFailed
	}
	return nil
}

// 用 sha256 哈希算法生成 JWT 签名, 传入 JWT Token 的前两部分和密钥, 返回生成的签名字符串
func signSha256(jwtMsg string, secret string) string {
	hash := hmac.New(sha256.New, []byte(secret))               // 根据 secret 生成 sha256 哈希算法器
	hash.Write([]byte(jwtMsg))                                 // 将 jwtMsg 写入
	return base64.RawURLEncoding.EncodeToString(hash.Sum(nil)) // 对哈希结果进行 base64 编码
}
