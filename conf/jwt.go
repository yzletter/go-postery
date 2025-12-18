package conf

const (
	RefreshTokenInCookie   = "refresh-token" // RefreshTokenMaxAgeSecs 在 cookie 中的 name
	RefreshTokenMaxAgeSecs = 5 * 86400
	AccessTokenExpiration  = 60 * 60
	RefreshTokenPrefix     = "auth:refresh:"
	ClearTokenPrefix       = "auth:clear:"
)
