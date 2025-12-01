package service

import (
	"errors"
	"time"

	"github.com/yzletter/go-postery/infra/viper"
)

var (
	JWTConfig = viper.InitViper("./conf", "jwt", viper.YAML)
)

// 可容忍的时间偏移
const defaultLeeway = 5 * time.Second

type JwtHeader struct {
	Algo string `json:"alg"` // 哈希算法, HS256
	Type string `json:"typ"` // JWT
}

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

var (
	ErrJwtInvalidParam       = errors.New("jwt 传入非法参数")
	ErrJwtMarshalFailed      = errors.New("jwt json 序列化失败")
	ErrJwtBase64DecodeFailed = errors.New("jwt json base64 解码失败")
	ErrJwtUnMarshalFailed    = errors.New("jwt json 反序列化失败")
	ErrJwtInvalidTime        = errors.New("jwt 时间错误")
)
