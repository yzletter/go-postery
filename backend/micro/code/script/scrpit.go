package script

import _ "embed"

const (
	// AllowCodeResultAbnormal 表示 Redis key 存在但没有过期时间，属于异常 key
	AllowCodeResultAbnormal = -1
	// AllowCodeResultTooFrequent 表示已发送验证码且未超过发送间隔，拒绝发送
	AllowCodeResultTooFrequent = 0
	// AllowCodeResultAllowed 表示未发送过验证码或已超过发送间隔，允许发送
	AllowCodeResultAllowed = 1
)

const (
	// VerifyCodeResultAbnormal 表示 Biz 异常
	VerifyCodeResultAbnormal = -1
	// VerifyCodeResultNotFound 表示验证码不存在或已过期
	VerifyCodeResultNotFound = 0
	// VerifyCodeResultMismatch 表示验证码错误
	VerifyCodeResultMismatch = 1
	// VerifyCodeResultMatched 表示验证码正确，脚本会删除 key 防止重复使用
	VerifyCodeResultMatched = 2
)

//go:embed lua/allow.lua
var AllowCodeScript string // 用于检查发送验证码的 Lua 脚本，返回 AllowCodeResult

//go:embed lua/verify.lua
var VerifyCodeScript string // 用于校验验证码的 Lua 脚本，返回 VerifyCodeResult
