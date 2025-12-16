package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

var (
	ErrNotLogin = errors.New("请先登录")
)

func GetUidFromCTX(ctx *gin.Context) (int64, error) {
	idStr, ok := ctx.Value(UID_IN_CTX).(string)
	if !ok {
		return 0, ErrNotLogin
	}

	uid, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		// 没有登录
		return 0, ErrNotLogin
	}

	return uid, nil
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
