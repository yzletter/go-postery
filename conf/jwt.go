package conf

const (
	AccessTokenInHeader          = "x-jwt-token"   // AccessToken 在 Header 中的 name
	RefreshTokenInCookie         = "refresh-token" // RefreshTokenCookieMaxAgeSecs 在 cookie 中的 name
	RefreshTokenCookieMaxAgeSecs = 5 * 86400
	AccessTokenExpiration        = 60 * 60
	RefreshTokenPrefix           = "auth:refresh:"
	ClearTokenPrefix             = "auth:clear:"
)
