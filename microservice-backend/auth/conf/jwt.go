package conf

const (
	RefreshTokenInCookie   = "refresh-token" // RefreshToken 在 cookie 中的 name
	AccessTokenInCookie    = "x-jwt-token"   // AccessToken 在 cookie 中的 name 用于 WS
	RefreshTokenMaxAgeSecs = 5 * 86400
	AccessTokenExpiration  = 60 * 60
	RefreshTokenPrefix     = "auth:refresh:"
	ClearTokenPrefix       = "auth:clear:"
)

const (
	JwtTokenKey = "123456"
)
