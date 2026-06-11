# Auth 模块

## 接口

```go
// AuthService 用户认证服务
type AuthService interface {
	// LoginByPassword 手机号码 / 邮箱 + 密码登录
	LoginByPassword(ctx context.Context, identifier string, password string) (int64, error)

	// LoginByPhone 手机号码 + 验证码进行登录, 未注册的手机号码自动进行注册
	LoginByPhone(ctx context.Context, phone string, code string) (int64, error)

	// HasPassword 查询密码状态
	HasPassword(ctx context.Context, id int64) (bool, error)

	// SetPassword 初始化密码
	SetPassword(ctx context.Context, uid int64, code string, password string) error

	// UpdatePassword 更新密码
	UpdatePassword(ctx context.Context, uid int64, oldPassword string, newPassword string) error

	// GetAuthIdentityByUID 获取用户身份认证
	GetAuthIdentityByUID(ctx context.Context, id int64) (string, string, error)

	// IssueTokens 签发双 Token
	IssueTokens(ctx context.Context, uid int64, role int, userAgent string) (string, string, error)

	// ClearTokens 清除双 Token
	ClearTokens(ctx context.Context, accessToken string, refreshToken string) error

	// VerifyAccessToken 校验 AccessToken
	VerifyAccessToken(ctx context.Context, accessToken string) (*ports.JWTTokenClaims, error)

	// GetInfoByRefreshToken 根据 RefreshToken 获取用户信息, 用于重新签发双 Token
	GetInfoByRefreshToken(ctx context.Context, refreshToken string) (int64, int, string, error)

	// CheckBlackList 根据 SSID 检查黑名单, 检查用户是否被拉黑
	CheckBlackList(ctx context.Context, ssid string) (bool, error)
}

```

## 长短双 Token 机制
![长短双 Token.png](../imgs/%E9%95%BF%E7%9F%AD%E5%8F%8C%20Token.png)
#### 用户信息

主要包括 UserID，SSID，Role

#### AccessToken

AccessToken 为 JWT Token，放在请求头中

#### RefreshToken

RefreshToken 为 随机复杂字符串，通过 Set-Cookie 写入

### 鉴权

#### HTTP

1. 从请求头中 Authorization 拿出 AccessToken
2. 从 Cookie 中拿出 RefreshToken
3. 尝试通过 AccessToken 鉴权
   1. 验证 JWT 是否正确
   2. 验证黑名单，若被拉黑直接返回
4. 尝试通过 RefreshToken 鉴权
   1. 根据 RefreshToken 从 Redis 获取用户信息
   2. 验证黑名单，若被拉黑直接返回
   3. 清除旧的 RefreshToken，拉黑旧 AccessToken，生成新的双 Token
   4. 将新的 AccessToken 放进 Header，新的 RefreshToken 放进 Cookie

#### Websocket

1. 从请求头中 Authorization 拿出 AccessToken，拿不到说明当为浏览器 Websocket 请求
2. 从 Cookie 中拿出 AccessToken 和 RefreshToken
3. 进行鉴权

### 黑名单

用特定 prefix + RefreshToken + ssid 组成 Key 并把 Value 置空存在 Redis 中表示用户被拉黑

## 用户登录

### 用户身份认证

主要由 uid + type + identifier 组成，分别表示用户 ID，身份认证方式和字段，例如：

>uid ：12345678910
>
>type：邮箱
>
>identifier：yzletter@foxmail.com

### 用户密码

主要由 uid + password 组成，password 可为空（手机号注册的用户）

### 手机号 / 邮箱 + 密码登录

1. 根据 identifier 获取身份认证
2. 记录不存在则无法登录
3. 校验密码
4. 返回用户 ID

### 手机号 + 验证码登录，未注册的手机号自动注册

1. 校验验证码
2. 根据 type + identifier  获取身份认证
3. 记录不存在进行注册（事务写 event 表发消息异步注册其他服务），记录存在返回用户 ID



