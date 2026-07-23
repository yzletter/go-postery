package conf

import "time"

const (
	OutboxLockTime = 60
	OutboxInterval = 5 * time.Second
)

// 限流配置
const (
	RateLimitInterval = time.Minute
	RateLimitRate     = 100
)

// 验证码相关配置
const (
	// 短信验证码相关配置
	SendSMSInterval = 60            // 发送间隔
	SMSValidTime    = 300           // 有效时间
	PhoneCodePrefix = "phone:code:" // 前缀

	// 邮箱验证码相关配置
	SendEmailInterval = 60            // 发送间隔
	EmailValidTime    = 600           // 有效时间
	EmailCodePrefix   = "email:code:" // 前缀
)

// Auth 相关配置
const (
	JwtTokenKey            = "123456"        // JWT 密钥
	UserIDInContext        = "user_id"       // uid 在上下文中的 Name
	RefreshTokenInCookie   = "refresh-token" // RefreshToken 在 cookie 中的 name
	AccessTokenInCookie    = "x-jwt-token"   // AccessToken 在 cookie 中的 name 用于 WS
	RefreshTokenMaxAgeSecs = 5 * 86400
	AccessTokenExpiration  = 60 * 60
	RefreshTokenPrefix     = "auth:refresh:"
	ClearTokenPrefix       = "auth:clear:"
)

// Lottery
const (
	RocketLotteryTopic             = "GO_POSTERY_CANCEL_ORDER"
	RocketLotteryConsumerGroup     = "go_postery"
	RocketAwaitDuration            = 5 * time.Second
	RocketLotteryPayDelay          = 600
	RocketLotteryInvisibleDuration = 10 * time.Second
)
