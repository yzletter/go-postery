# go-postery

## 简介

用 Go 实现一个简单的信息发布系统

## 设计文档

[go-postery 设计文档](https://yzletter.notion.site/go-postery-2b389200bcae80ef93abeee61f52ea4c?source=copy_link)

## 一阶段文档

## Step 1. 设计数据库表

### 建库

- 数据库名: `go_postery`
- 数据库管理用户名: `go_postery_tester`
  - 密码: `123456`

```sql
# 建库
-- 创建数据库 go_postery
create database go_postery;
-- 创建用户 go_postery_tester 密码为 123456
create user 'go_postery_tester' identified by '123456';
-- 将数据库 go_postery 的全部权限授予用户 go_postery_tester
grant all on go_postery.* to go_postery_tester;
-- 切到 go_postery 数据库
use go_postery;
```

### 建表

创建 `user` 表

```sql
# 创建 user 表
create table if not exists user
(
    id          int auto_increment comment '用户 id, 自增',
    name        varchar(20) not null comment '用户名',
    password    char(32)    not null comment '用户密码的 md5 加密结果',
    create_time datetime default current_timestamp comment '用户注册时间, 默认为创建记录的时间',
    update_time datetime default current_timestamp on update current_timestamp comment '用户最后修改时间',
    primary key (id),
    unique key idx_name (name)
) default charset = utf8mb4 comment '用户信息表';
```

创建 `post` 表

```sql
# 创建 post 表
create table if not exists post
(
    id          int auto_increment comment '帖子 id, 自增',
    user_id     varchar(20) not null comment '发布者 id',
    create_time datetime default current_timestamp comment '帖子创建时间',
    update_time datetime default current_timestamp on update current_timestamp comment '帖子最后修改时间',
    delete_time datetime default null comment '帖子删除时间',
    title      varchar(100) comment '标题',
    content     text comment '正文',
    primary key (id),
    unique key idx_user (user_id)
) default charset = utf8mb4 comment '帖子信息表';
```

## Step 2. 初始化 `Viper`

### 编写 `InitViper` 函数

```go
// InitViper 初始化 Viper 读取配置, 传入目录、文件名、文件类型, 返回一个 viper.Viper 指针
func InitViper(dir, fileName, fileType string) *viper.Viper {
	config := viper.New()          // 创建 *Viper 对象
	config.AddConfigPath(dir)      // 设置配置文件所在目录
	config.SetConfigName(fileName) // 设置配置文件名 (无路径, 无后缀)
	config.SetConfigType(fileType) // 设置文件类型

	// 尝试解析配置文件
	if err := config.ReadInConfig(); err != nil {
		configFile := path.Join(dir, fileName) + "." + fileType // 完整配置文件路径
		// 系统初始化过程中发生错误直接 panic, logger 还未初始化, 不能用 logger.fatal()
		panic(fmt.Errorf("go-postery viper : 解析 [%s] 配置文件出错 > %s", configFile, err))
	}

	return config
}
```

### 修改项目目录结构

```bash
./
├── README.md
├── database
│   └── create_table.sql
├── go.mod
├── go.sum
└── utils
    └── viper.go
```

## Step 3. 连接到数据库

### 编写数据库配置文件 `db.yaml`

```yaml
mysql:
  host: localhost         # 地址
  port: 3306              # 端口号
  user: go_postery_tester # 用户名
  password: 123456        # 密码
  dbName: go_postery      # 数据库名
  logFileName: db.log     # 日志文件名
```

### 编写 `ConnectToDB` 连接到数据库

```go
var (
	GoPosteryDB *gorm.DB // 定义全局数据库变量 GoPosteryDB
)

// ConnectToDB 连接到 MySQL 数据库, 生成一个 *gorm.DB 赋给全局数据库变量 GoPosteryDB
func ConnectToDB(confDir, confFileName, confFileType, logDir string) {
	// 读取 MySQL 相关配置
	viper := utils.InitViper(confDir, confFileName, confFileType) // 初始化一个 Viper 进行配置读取
	host := viper.GetString("mysql.host")
	port := viper.GetInt("mysql.port")
	user := viper.GetString("mysql.user")
	password := viper.GetString("mysql.password")
	dbName := viper.GetString("mysql.dbName")
	logFileName := viper.GetString("mysql.logFileName")

	// 拼接完整的请求路径 user:password@tcp(host:port)/dbName?charset=utf8mb4&parseTime=True&loc=local
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=local", user, password, host, port, dbName)

	// 设置 logger 相关配置
	logFile, err := os.OpenFile(path.Join(logDir, logFileName), os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm) // 打开日志文件
	if err != nil {
		panic(fmt.Errorf("go-postery ConnectToDB : 打开目标 logger 文件失败 %s", err))
	}
	loggerConfig := logger.Config{
		SlowThreshold:             100 * time.Millisecond, // 超过此阈值为慢查询
		Colorful:                  false,                  // 禁用颜色, 可提高性能
		IgnoreRecordNotFoundError: true,                   // 忽略 RecordNotFound 这种错误日志
		LogLevel:                  logger.Info,            // 日志最低阈值
	}
	myDBLogger := logger.New(
		log.New(logFile, "\\r\\n", log.LstdFlags), // io writer，可以输出到文件，也可以输出到os.Stdout
		loggerConfig,
	)

	// 设置 gorm 相关配置
	gormConfig := &gorm.Config{
		PrepareStmt:            true, // 执行任一 SQL 语句时, 都会创建 Prepare Statement 并缓存, 以提高后续执行效率
		SkipDefaultTransaction: true, // 禁止在事务中进行写入操作, 性能提升约 30%
		// 覆盖默认命名策略
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 表名映射不加复数, 仅仅是驼峰转为蛇形
		},
		Logger: myDBLogger, // 日志控制
	}

	// 建立连接
	db, err := gorm.Open(mysql.Open(dataSourceName), gormConfig) // 生成一个 *gorm.DB
	if err != nil {
		panic(fmt.Errorf("go-postery ConnectToDB : 连接到数据库出错 %s", err))
	}

	// Mysql 连接池参数配置
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(10)  // 连接池中空闲连接数目上限, 超出此上限就把相应的连接关闭掉
	sqlDB.SetMaxOpenConns(100) // 最多打开链接数目上限
	// 单个连接的连接时间上限, 超时会自动关闭, 因为数据库本身可能也对NoActive连接设置了超时时间, 我们的应对办法: 定期 ping, 或者 SetConnMaxLifetime
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 赋给全局变量 GoPosteryDB
	GoPosteryDB = db
}
```

### 测试数据库连接

```go
func TestConnection(t *testing.T) {
	database.ConnectToDB("../../conf", "db", "yaml", "../../log")
	sqlDB, err := database.GoPosteryDB.DB()
	if err != nil {
		t.Fatalf("获取 sql.DB 失败: %v", err)
	}
	err = sqlDB.Ping()
	if err != nil {
		t.Fatalf("Ping 失败: %v", err)
	}
}

// go test -v ./database/gorm -run=^TestConnection$ -count=1
```

### 修改项目目录结构

```bash
./
├── README.md
├── conf
│   └── db.yaml
├── database
│   ├── create_table.sql
│   └── gorm
│       ├── connection.go
│       └── connection_test.go
├── go.mod
├── go.sum
├── log
│   └── db.log
└── utils
    └── viper.go
```

## Step 4. 初始化 Slog

### 编写 `InitSlog` 函数

```go
// InitSlog 初始化 Slog
func InitSlog(logFileName string) {
	// 设置 rotatelogs 滚动日志相关配置
	logFile, err := rotatelogs.New(
		logFileName+".%Y%m%d%H",                  // 日志文件路径
		rotatelogs.WithLinkName(logFileName),     // 创建软链接指向最新的一份日志
		rotatelogs.WithRotationTime(1*time.Hour), // 设置滚动时间, 每小时滚动一次
		rotatelogs.WithMaxAge(7*24*time.Hour),    // 设置日志保存时间, 或使用 WithRotationCount 只保留最近的几份日志
	)

	if err != nil {
		panic(err)
	}

	// 设置 Slog 相关配置
	slogConfig := &slog.HandlerOptions{
		AddSource: true,           // 上报文件名和行号
		Level:     slog.LevelInfo, // 设置日志最低级别
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr { // 用 Go 标准时间格式替换默认时间格式
			if a.Key == slog.TimeKey { // 如果 Key == "time"
				t := a.Value.Time()
				a.Value = slog.StringValue(t.Format("2006-01-02 15:04:05.000")) // 替换 Value
			}
			return a
		},
	}

	// 构造 logger
	slogHandler := slog.NewTextHandler( // JSON 格式
		logFile,    // 指定文件
		slogConfig, // 相关配置
	)
	logger := slog.New(slogHandler)

	// 设置为全局 logger
	slog.SetDefault(logger)
}
```

## Step 5. 用户注册

### 模型映射

```go
// User 定义数据库中 user 表的模型映射
type User struct {
	Id       int    `gorm:"primaryKey"`      // 用户 ID
	Name     string `gorm:"column:name"`     // 用户名
	PassWord string `gorm:"column:password"` // 用户密码 MD5 后的结果
}
```

### 编写 `RegisterUser` 函数

```go
// RegisterUser 传入 name 和 password 注册新用户, 返回 Id 和可能的错误
func RegisterUser(name, password string) (int, error) {
	// 将模型绑定到结构体
	user := model.User{
		Name:     name,
		PassWord: password,
	}

	// 到 MySQL 中创建新记录
	err := GoPosteryDB.Create(&user).Error // 需要传指针
	// 错误处理
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) { // 判断是否为 MySQL 错误
			if mysqlErr.Number == 1062 { // Unique Key 冲突
				return 0, fmt.Errorf("用户[%s]已存在", name)
			}
		}
		// 记录日志, 方便后续人工定位问题所在
		slog.Error("go-postery RegisterUser : 用户注册失败", "name", name, "error", err)
		return 0, fmt.Errorf("用户注册失败，请稍后重试")
	}

	// 返回 Id
	return user.Id, nil
}
```

### 测试用户注册函数

```go
package database_test

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"testing"

	database "github.com/yzletter/go-postery/database/gorm"
	"github.com/yzletter/go-postery/utils"
)

// Hash 返回字符串 MD5 哈希后 32 位的十六进制编码结果
func hash(password string) string {
	hasher := md5.New()
	hasher.Write([]byte(password))
	digest := hasher.Sum(nil)
	return hex.EncodeToString(digest)
}

func init() {
	utils.InitSlog("../../log/go_postery.log")
	database.ConnectToDB("../../conf", "db", "yaml", "../../log")
}

func TestRegisterUser(t *testing.T) {
	// 注册一次 yzletter, 结果应为成功
	id1, err := database.RegisterUser("yzletter", hash("123456"))
	if err != nil {
		fmt.Printf("用户[%d]注册失败 \\n", id1)
		t.Fatal()
	} else {
		fmt.Printf("用户[%d]注册成功 \\n", id1)
	}

	// 再注册一次 yzletter, 结果应为失败
	id2, err := database.RegisterUser("yzletter", hash("123456"))
	if err == nil {
		fmt.Printf("用户[%d]重复成功 \\n", id2)
		t.Fatal()
	} else {
		fmt.Println("用户重复注册")
	}
}

// go test -v ./database/gorm -run=^TestRegisterUser$ -count=1
```

## Step 6. 用户管理

### `UpdatePassword`更改用户密码

```go
// UpdatePassword 根据传入的用户 id, 新旧密码更改用户密码
func UpdatePassword(uid int, oldPass, newPass string) error {
	tx := GoPosteryDB.Model(&model.User{}).Where("id=? and password=?", uid, oldPass).Update("password", newPass)
	if tx.Error != nil {
		// 系统错误
		slog.Error("go-postery UpdatePassword : 密码更改失败", "uid", uid, "error", tx.Error)
		return fmt.Errorf("更改用户密码失败, 请稍后再试")
	} else if tx.RowsAffected == 0 {
		// 业务错误
		return fmt.Errorf("用户 id 或旧密码错误")
	}

	return nil
}
```

`GetUserById`根据 Id 查找用户

```go
// GetUserById 根据 Id 查找用户
func GetUserById(uid int) *model.User {
	user := model.User{Id: uid}
	tx := GoPosteryDB.Select("*").First(&user) // 隐含的where条件是id, 注意：Find不会返回ErrRecordNotFound
	if tx.Error != nil {
		// 若错误不是记录未找到, 记录系统错误
		if !errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			slog.Error("go-postery GetUserById : 查找用户失败", "uid", uid, "error", tx.Error)
		}
		return nil
	}

	return &user
}
```

`LogOffUser`根据传入的 Id 注销用户

```go
// LogOffUser 根据传入的 id 注销用户
func LogOffUser(uid int) error {
	// 将模型绑定到结构体
	user := model.User{
		Id: uid,
	}
	// 删除记录
	tx := GoPosteryDB.Delete(&user)
	if tx.Error != nil {
		// 系统层面错误
		slog.Error("go-postery LogOffUser : 用户注销失败", "uid", uid, "error", tx.Error)
		return fmt.Errorf("用户注销失败，请稍后重试")
	} else if tx.RowsAffected == 0 {
		// 业务层面错误
		return fmt.Errorf("用户注销失败, uid %d 不存在", uid)
	}
	return nil
}
```

### 测试函数

```go
func TestLogOffUser(t *testing.T) {
	var uid = 7
	// 删前查询
	user := database.GetUserById(uid)
	fmt.Println(user)

	err := database.LogOffUser(uid)
	if err != nil {
		t.Fatal(err)
	} else {
		fmt.Println("首次删除成功")
	}

	// 删完后再查询
	user = database.GetUserById(uid)
	fmt.Println(user)

	// 再删一次
	err = database.LogOffUser(uid)
	if err == nil {
		t.Fatal(err)
	} else {
		fmt.Println("重复删除失败")
	}
}

func TestUpdatePassword(t *testing.T) {
	var uid = 9
	err := database.UpdatePassword(uid, hash("123456"), hash("654321"))
	if err != nil {
		t.Fatal(err)
	}
}
```

## Step 7. 注册、登录、登出、修改密码

### 模型映射

```go
package model

// User 定义前端提交登录表单信息的模型映射
type User struct {
	Name     string `form:"name" binding:"required, gte=2"`      // 长度 >= 2
	PassWord string `form:"password" binding:"required, len=32"` // 长度 == 32
}

// ModifyPasswordRequest 定义前端提交修改密码表单信息的模型映射
type ModifyPasswordRequest struct {
	OldPass string `form:"old_pass" binding:"required, len=32"` // 长度 == 32
	NewPass string `form:"new_pass" binding:"required, len=32"` // 长度 == 32
}
```

### `LoginHandlerFunc`

```go
// LoginHandlerFunc 用户登录 Handler
func LoginHandlerFunc(ctx *gin.Context) {
	var loginRequest = model.LoginRequest{}
	// 将请求参数绑定到结构体
	err := ctx.ShouldBind(&loginRequest)
	if err != nil {
		// 参数绑定失败
		ctx.String(http.StatusBadRequest, "用户名或密码错误")
		return
	}
	user := database.GetUserByName(loginRequest.Name)
	if user == nil {
		// 根据 name 未找到 user
		ctx.String(http.StatusBadRequest, "用户名或密码错误")
		return
	}
	if user.PassWord != loginRequest.PassWord {
		// 密码不正确
		ctx.String(http.StatusBadRequest, "用户名或密码错误")
		return
	}

	// 设置 Cookie
	ctx.SetCookie("uid", strconv.Itoa(user.Id), 86400, "/", "localhost", false, true)

	// 默认情况下也返回200
	ctx.String(http.StatusOK, "登录成功")
}
```

### `LogoutHandlerFunc`

```go
// LogoutHandlerFunc 用户登出 Handler
func LogoutHandlerFunc(ctx *gin.Context) {
	// 设置 Cookie
	ctx.SetCookie("uid", "", -1, "/", "localhost", false, true)
}
```

### `ModifyPassHandlerFunc`

```go
// ModifyPassHandlerFunc 修改密码 Handler
func ModifyPassHandlerFunc(ctx *gin.Context) {
	var modifyPassRequest model.ModifyPasswordRequest
	// 将请求参数绑定到结构体
	err := ctx.ShouldBind(&modifyPassRequest)
	if err != nil {
		// 参数绑定失败
		ctx.String(http.StatusBadRequest, "密码输入错误")
		return
	}
	uid := GetUidFromCookie(ctx)
	if uid == 0 {
		// 没有登录
		ctx.String(http.StatusBadRequest, "请先登录")
		return
	}

	err = database.UpdatePassword(uid, modifyPassRequest.OldPass, modifyPassRequest.NewPass)
	if err != nil {
		// 密码更改失败
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	// 默认情况下也返回200
	ctx.String(http.StatusOK, "密码修改成功")
}
```

### `RegisterHandlerFunc`

```go
// RegisterHandlerFunc 用户注册 Handler
func RegisterHandlerFunc(ctx *gin.Context) {
	var createUserRequest model.RegisterRequest
	err := ctx.ShouldBind(&createUserRequest)
	if err != nil {
		// 参数绑定失败
		ctx.String(http.StatusBadRequest, "用户名或密码错误")
		return
	}

	_, err = database.RegisterUser(createUserRequest.Name, createUserRequest.PassWord)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}
}
```

### `main`函数

```go
package main

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	database "github.com/yzletter/go-postery/database/gorm"
	handler "github.com/yzletter/go-postery/handler/gin"
	"github.com/yzletter/go-postery/utils"
)

func main() {
	// 初始化
	utils.InitSlog("./log/go_postery.log")
	database.ConnectToDB("./conf", "db", "yaml", "./log")

	r := gin.Default()

	// 配置跨域
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // 允许所有域名跨域
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 定义路由
	r.POST("/register", handler.RegisterHandlerFunc)     // 用户注册
	r.POST("/login", handler.LoginHandlerFunc)           // 用户登录
	r.GET("/logout", handler.LogoutHandlerFunc)          // 用户退出
	r.POST("/modifypass", handler.ModifyPassHandlerFunc) // 修改密码
	if err := r.Run("127.0.0.1:8080"); err != nil {
		panic(err)
	}
}
```

## Step 8. JWT

### 编写 `JWT` 模块

JWT 生成和校验函数

```go
package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

var (
	// DefaultHeader 默认的 JWT Header
	DefaultHeader = JwtHeader{
		Algo: "HS256",
		Type: "JWT",
	}
	ErrJwtInvalidParam       = errors.New("jwt 传入非法参数")
	ErrJwtMarshalFailed      = errors.New("jwt json 序列化失败")
	ErrJwtBase64DecodeFailed = errors.New("jwt json base64 解码失败")
	ErrJwtUnMarshalFailed    = errors.New("jwt json 反序列化失败")
	ErrJwtInvalidTime        = errors.New("jwt 时间错误")
)

// 可容忍的时间漂移
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

// GenJWT 根据 payload 和 secret 生成 JWT, 返回生成的 JWT token 和可能的错误
func GenJWT(payload JwtPayload, secret string) (string, error) {
	// 参数校验
	if secret == "" {
		return "", ErrJwtInvalidParam
	}

	// 1. header 转成 json, 再用 base64 编码, 得到 JWT 第一部分
	header := DefaultHeader
	part1, err := marshalBase64Encode(header)
	if err != nil {
		return "", err
	}

	// 2. payload 转成 json, 再用 base64 编码, 得到 JWT 第二部分
	part2, err := marshalBase64Encode(payload)
	if err != nil {
		return "", err
	}

	// 3. 根据 msg 使用 secret 进行加密得到签名 signature
	jwtMsg := part1 + "." + part2              // JWT 信息部分
	jwtSignature := signSha256(jwtMsg, secret) // JWT 签名部分

	return jwtMsg + "." + jwtSignature, nil
}

// VerifyJWT 根据传入的 JWT token 和 secret 校验 JWT 的合法性
func VerifyJWT(jwtToken string, secret string) (*JwtPayload, error) {
	// 参数校验
	if jwtToken == "" || secret == "" {
		return nil, ErrJwtInvalidParam
	}
	parts := strings.SplitN(jwtToken, ".", 3)
	if len(parts) != 3 {
		// 传入的 JWT 格式有误
		return nil, ErrJwtInvalidParam
	}

	// 获得 msg 和 signature 部分
	jwtMsg := parts[0] + "." + parts[1]
	jwtSignature := parts[2]

	// 1. 签名校验
	// 对 jwtMsg 加密得到 thisSignature 判断与 jwtSignature 是否相同
	thisSignature := signSha256(jwtMsg, secret)
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
```

JWT 配置

```yaml
secret: 123456
```

### 修改 `LoginHandlerFunc`

```go
// LoginHandlerFunc 用户登录 Handler
func LoginHandlerFunc(ctx *gin.Context) {
	var loginRequest = model.LoginRequest{}
	// 将请求参数绑定到结构体
	err := ctx.ShouldBind(&loginRequest)
	if err != nil {
		// 参数绑定失败
		ctx.String(http.StatusBadRequest, "用户名或密码错误")
		return
	}
	user := database.GetUserByName(loginRequest.Name)
	if user == nil {
		// 根据 name 未找到 user
		ctx.String(http.StatusBadRequest, "用户名或密码错误")
		return
	}
	if user.PassWord != loginRequest.PassWord {
		// 密码不正确
		ctx.String(http.StatusBadRequest, "用户名或密码错误")
		return
	}

	slog.Info("登录成功", "uid", user.Id)

	// 使用 JWT
	payload := utils.JwtPayload{
		Issue:       "yzletter",
		IssueAt:     time.Now().Unix(),                              // 签发日期为当前时间
		Expiration:  time.Now().Add(86400 * 7 * time.Second).Unix(), // 7 天后过期
		UserDefined: map[string]any{"uid": user.Id},                 // 用户自定义字段
	}

	jwtToken, err := utils.GenJWT(payload, "123456")
	if err != nil {
		// jwt 生成失败
		slog.Error("jwt 生成失败", "error", err)
		ctx.String(http.StatusInternalServerError, "jwt 生成失败")
	} else {
		// 生成成功, 放入 Cookies
		ctx.SetCookie("jwt", jwtToken, 86400*7, "/", "localhost", false, true)
	}

	// 默认情况下也返回200
	ctx.String(http.StatusOK, "登录成功")
}
```

### 身份认证中间件 `Auth`

```go
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yzletter/go-postery/utils"
)

const (
	UID_IN_JWT      = "uid" // uid 在 jwt 自定义字段的 name
	UID_IN_CTX      = "uid" // uid 在上下文中的 name
	JWT_COOKIE_NAME = "jwt" // jwt 在 cookie 中的 name
)

var (
	JWTConfig = utils.InitViper("./conf", "jwt", utils.YAML)
)

// AuthHandlerFunc 身份认证中间件
func AuthHandlerFunc(ctx *gin.Context) {
	// 获取登录 uid
	jwtToken := getJWTFromCookie(ctx)
	uid := getUidFromJWT(jwtToken)

	// 判断 uid 是否合法
	if uid == 0 {
		ctx.Redirect(http.StatusTemporaryRedirect, "/login") // 未登录, 进行重定向
		ctx.Abort()                                          // 当前中间件执行完, 后续中间件不执行
		return
	}

	// 把 uid 放入上下文, 以便后续中间件直接使用
	ctx.Set(UID_IN_CTX, uid)
}

// 从 cookie 中获取 JWT Token
func getJWTFromCookie(ctx *gin.Context) string {
	jwtToken := ""
	for _, cookie := range ctx.Request.Cookies() {
		if cookie.Name == JWT_COOKIE_NAME {
			jwtToken = cookie.Value
			break
		}
	}
	return jwtToken
}

// 从 JWT Token 中获取 uid
func getUidFromJWT(jwtToken string) int {
	payload, err := utils.VerifyJWT(jwtToken, JWTConfig.GetString("secret")) // 加密的 key 从配置文件中读取
	if err != nil {
		return 0 // jwt 校验失败
	}
	for k, v := range payload.UserDefined {
		if k == UID_IN_JWT {
			return int(v.(float64)) // Json 反序列化 map[string]any 时，数字会被解析成 float64，而不是 int
		}
	}
	return 0 // 未找到 uid
}
```

## Step 9. 帖子模块开发

### 模型映射

```go
package model

import "time"

type Post struct {
	Id         int        `gorm:"primaryKey"`
	UserId     int        `gorm:"column:user_id"`
	Title     string     `gorm:"column:title"`
	Content    string     `gorm:"column:content"`
	CreateTime *time.Time `gorm:"column:create_time"`
	DeleteTime *time.Time `gorm:"column:delete_time"`
	ViewTime   string     `gorm:"-"` // 前端用于展示的时间
}
```

### 数据库层函数

```go
package database

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/yzletter/go-postery/database/model"
	"gorm.io/gorm"
)

// CreatePost 新建帖子
func CreatePost(uid int, title, content string) (int, error) {
	// 模型映射
	now := time.Now()
	post := model.Post{
		UserId:     uid,     // 作者id
		Title:     title,  // 标题
		Content:    content, // 正文
		CreateTime: &now,
		DeleteTime: nil, // 须显示指定为 nil, 写入数据库为 null,
	}

	// 新建数据
	if err := GoPosteryDB.Create(&post).Error; err != nil {
		slog.Error("帖子发布失败", "title", err)
		return 0, errors.New("帖子发布失败")
	}

	return post.Id, nil
}

// DeletePost 根据帖子 id 删除帖子
func DeletePost(pid int) error {
	tx := GoPosteryDB.Model(&model.Post{}).Where("id=? and delete_time is null", pid).Update("delete_time", time.Now())
	if tx.Error != nil {
		// 删除失败
		slog.Error("帖子删除失败", "pid", pid, "error", tx.Error)
		return errors.New("帖子删除失败")
	} else {
		if tx.RowsAffected == 0 {
			return fmt.Errorf("帖子 %d 不存在", pid)
		} else {
			return nil
		}
	}
}

// UpdatePost 修改帖子
func UpdatePost(pid int, title, content string) error {
	tx := GoPosteryDB.Model(&model.Post{}).Where("id=? and delete_time is null", pid).
		Updates(map[string]interface{}{
			"title":  title,
			"content": content,
		})
	if tx.Error != nil {
		//	修改失败
		slog.Error("帖子修改失败", "pid", pid, "error", tx.Error)
		return errors.New("帖子修改失败")
	} else {
		if tx.RowsAffected == 0 {
			return fmt.Errorf("帖子 %d 不存在", pid)
		} else {
			return nil
		}
	}
}

// GetPostByID 根据帖子 id 获取帖子信息
func GetPostByID(pid int) *model.Post {
	post := &model.Post{
		Id: pid,
	}
	tx := GoPosteryDB.Select("*").Where("delete_time is null").First(post) // find 不会报 ErrNotFound
	if tx.Error != nil {
		if !errors.Is(tx.Error, gorm.ErrRecordNotFound) { // 并非未找到, 而是其他错误
			slog.Error("帖子查找失败", "pid", pid, "error", tx.Error)
		}
		return nil
	}

	// 赋值前端用于显示的时间
	post.ViewTime = post.CreateTime.Format("2006-01-02 15:04:05")
	return post
}

// GetPostByPage 翻页查询帖子, 页号从 1 开始, 返回帖子总数和帖子列表
func GetPostByPage(pageNumber, pageSize int) (int, []*model.Post) {
	// 获取帖子总数
	var total int64
	tx := GoPosteryDB.Model(&model.Post{}).Where("delete_time is null").Count(&total)
	if tx.Error != nil {
		slog.Error("获取帖子总数失败", "error", tx.Error)
		return 0, nil
	}

	// 获取当前页的帖子
	var posts []*model.Post

	// 已经查询过 pageSize * (pageNumber - 1) 条数据, 当前页需要 pageSize 条数据，并按发布时间降序排列
	tx = GoPosteryDB.Model(&model.Post{}).Where("delete_time is null").Order("create_time desc").Limit(pageSize).Offset(pageSize * (pageNumber - 1)).Find(&posts)
	if tx.Error != nil {
		slog.Error("获取当前页帖子失败", "pageNumber", pageNumber, "pageSize", pageSize, "error", tx.Error)
		return 0, nil
	}

	// 赋值前端展示时间
	for _, post := range posts {
		post.ViewTime = post.CreateTime.Format("2006-01-02 15:04:05")
	}

	return int(total), posts
}

// GetPostByUid 根据 uid 获取该用户所发帖子
func GetPostByUid(uid int) []*model.Post {
	// todo
	return nil
}
```

### `handler`开发

```go
package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	database "github.com/yzletter/go-postery/database/gorm"
	"github.com/yzletter/go-postery/handler/model"
	"github.com/yzletter/go-postery/utils"
)

// GetPostsHandler 获取帖子列表
func GetPostsHandler(ctx *gin.Context) {
	// 从 /posts?pageNo=1&pageSize=2 路由中拿出 pageNo 和 pageSize
	pageNo, err1 := strconv.Atoi(ctx.DefaultQuery("pageNo", "1"))
	pageSize, err2 := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))
	if err1 != nil || err2 != nil {
		res := utils.Resp{
			Code: 1,
			Msg:  "获取帖子列表请求的参数不合法",
			Data: nil,
		}
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	// 获取帖子总数和当前页帖子列表
	total, posts := database.GetPostByPage(pageNo, pageSize)
	postsBack := []gin.H{}
	for _, post := range posts {
		// 根据 uid 找到 username 进行赋值
		user := database.GetUserById(post.UserId)
		if user != nil {
			post.UserName = user.Name
		} else {
			slog.Warn("could not get name of user", "uid", post.UserId)
		}

		res := gin.H{
			"id":      post.Id,
			"title":   post.Title,
			"content": post.Content,
			"author": gin.H{
				"id":   post.UserId,
				"name": post.UserName,
			},
			"createdAt": post.ViewTime,
		}
		postsBack = append(postsBack, res)
	}

	// 计算是否还有帖子 = 判断已经加载的帖子数是否小于总帖子数
	hasMore := pageNo*pageSize < total

	resp := utils.Resp{
		Code: 0,
		Msg:  "获取帖子列表成功",
		Data: gin.H{
			"posts":   postsBack,
			"total":   total,
			"hasMore": hasMore,
		},
	}
	ctx.JSON(http.StatusOK, resp)
	return
}

// GetPostDetailHandler 获取帖子详情
func GetPostDetailHandler(ctx *gin.Context) {
	// 从路由中获取 pid 参数
	pid, err := strconv.Atoi(ctx.Param("pid"))
	if err != nil {
		resp := utils.Resp{
			Code: 1,
			Msg:  "获取帖子详情失败",
			Data: nil,
		}
		ctx.JSON(http.StatusOK, resp)
	}

	// 根据 pid 查找帖子详情
	post := database.GetPostByID(pid)
	if post == nil {
		resp := utils.Resp{
			Code: 1,
			Msg:  "获取帖子详情失败",
			Data: nil,
		}
		ctx.JSON(http.StatusOK, resp)
		return
	}

	// 获取作者用户名
	user := database.GetUserById(post.UserId)
	if user != nil {
		post.UserName = user.Name
	} else {
		slog.Warn("could not get name of user", "uid", post.UserId)
	}

	resp := utils.Resp{
		Code: 0,
		Msg:  "获取帖子详情成功",
		Data: gin.H{
			"id":      post.Id,
			"title":   post.Title,
			"content": post.Content,
			"author": gin.H{
				"id":   post.UserId,
				"name": post.UserName,
			},
			"createdAt": post.ViewTime,
		},
	}
	ctx.JSON(http.StatusOK, resp)
}

// CreateNewPostHandler 创建帖子
func CreateNewPostHandler(ctx *gin.Context) {
	// 直接从 ctx 中拿 uid
	uid := ctx.Value(UID_IN_CTX).(int)

	// 参数绑定
	var createRequest model.CreateRequest
	err := ctx.ShouldBind(&createRequest)
	if err != nil {
		resp := utils.Resp{
			Code: 1,
			Msg:  "创建帖子参数错误",
			Data: nil,
		}
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}

	// 创建帖子
	pid, err := database.CreatePost(uid, createRequest.Title, createRequest.Content)
	if err != nil {
		// 创建帖子失败
		resp := utils.Resp{
			Code: 1,
			Msg:  "创建帖子失败,请稍后重试",
			Data: nil,
		}
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}

	resp := utils.Resp{
		Code: 0,
		Msg:  "创建帖子成功",
		Data: gin.H{
			"id": pid,
		},
	}
	ctx.JSON(http.StatusOK, resp)
}

// DeletePostHandler 删除帖子
func DeletePostHandler(ctx *gin.Context) {
	// 直接从 ctx 中拿 uid
	uid := ctx.Value(UID_IN_CTX).(int)

	// 再拿帖子 pid
	pid, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || pid == 0 {
		resp := utils.Resp{
			Code: 1,
			Msg:  "帖子 id 获取失败",
		}
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}

	// 判断登录用户是否是作者
	post := database.GetPostByID(pid)
	if post == nil {
		resp := utils.Resp{
			Code: 1,
			Msg:  "当前帖子不存在",
		}
		ctx.JSON(http.StatusBadRequest, resp)
		return
	} else if uid != post.UserId {
		// 无权限删除
		resp := utils.Resp{
			Code: 1,
			Msg:  "无权限删除该帖子",
		}
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}

	// 进行删除
	err = database.DeletePost(pid)
	if err != nil {
		resp := utils.Resp{
			Code: 1,
			Msg:  "帖子删除失败",
		}
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}

	resp := utils.Resp{
		Code: 0,
		Msg:  "帖子删除成功",
	}
	ctx.JSON(http.StatusOK, resp)
	return
}

// UpdatePostHandler 修改帖子
func UpdatePostHandler(ctx *gin.Context) {
	// 直接从 ctx 中拿 uid
	uid := ctx.Value(UID_IN_CTX).(int)

	// 参数绑定
	var updateRequest model.CreateRequest
	err := ctx.ShouldBind(&updateRequest)
	if err != nil || updateRequest.Id == 0 {
		resp := utils.Resp{
			Code: 1,
			Msg:  "修改帖子参数错误",
		}
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}

	// 判断登录用户是否是作者
	post := database.GetPostByID(updateRequest.Id)
	if post == nil {
		resp := utils.Resp{
			Code: 1,
			Msg:  "当前帖子不存在",
		}
		ctx.JSON(http.StatusBadRequest, resp)
		return
	} else if uid != post.UserId {
		// 无权限删除
		resp := utils.Resp{
			Code: 1,
			Msg:  "无权限修改该帖子",
		}
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}

	// 修改
	err = database.UpdatePost(updateRequest.Id, updateRequest.Title, updateRequest.Content)
	if err != nil {
		resp := utils.Resp{
			Code: 1,
			Msg:  "修改失败，请稍后重试",
		}
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}

	resp := utils.Resp{
		Code: 0,
		Msg:  "帖子修改成功",
	}
	ctx.JSON(http.StatusOK, resp)
	return
}

// PostBelongHandler 查询帖子作者是否为当前登录用户
func PostBelongHandler(ctx *gin.Context) {
	// 获取帖子 id
	pid, err := strconv.Atoi(ctx.Query("id"))
	if err != nil {
		resp := utils.Resp{
			Code: 0,
			Msg:  "帖子不属于当前用户",
			Data: "false",
		}
		ctx.JSON(http.StatusOK, resp)
		return
	}

	// 获取登录 uid
	jwtToken := getJWTFromCookie(ctx)
	uid := getUidFromJWT(jwtToken)
	slog.Info("Auth", "uid", uid)

	if uid == 0 {
		// 未登录, 后面不用看了
		resp := utils.Resp{
			Code: 0,
			Msg:  "帖子不属于当前用户",
			Data: "false",
		}
		ctx.JSON(http.StatusOK, resp)
		return
	}

	// 判断登录用户是否是作者
	post := database.GetPostByID(pid)
	if post == nil || uid != post.UserId {
		resp := utils.Resp{
			Code: 0,
			Msg:  "帖子不属于当前用户",
			Data: "false",
		}
		ctx.JSON(http.StatusOK, resp)
		return
	}

	// 属于
	resp := utils.Resp{
		Code: 0,
		Msg:  "帖子属于当前用户",
		Data: "true",
	}
	ctx.JSON(http.StatusOK, resp)
	return
}
```

### 路由定义

```go
	// 帖子模块
	engine.GET("/posts", handler.GetPostsHandler)                                       // 获取帖子列表
	engine.GET("/posts/:pid", handler.GetPostDetailHandler)                             // 获取帖子详情
	engine.POST("/posts/new", handler.AuthHandlerFunc, handler.CreateNewPostHandler)    // 创建帖子
	engine.GET("/posts/delete/:id", handler.AuthHandlerFunc, handler.DeletePostHandler) // 删除帖子
	engine.POST("/posts/update", handler.AuthHandlerFunc, handler.UpdatePostHandler)    // 修改帖子
	engine.GET("/posts/belong", handler.PostBelongHandler)                              // 查询帖子是否归属当前登录用户
```

## 定时任务

ping

```go
func Ping() {
	if GoPosteryDB != nil {
		sqlDB, _ := GoPosteryDB.DB()
		err := sqlDB.Ping()
		if err != nil {
			slog.Info("ping GoPosteryDB failed")
			return
		}
		slog.Info("ping GoPosteryDB succeed")
		return
	}
}
```

### `crontab`

```go
func InitCrontab() {
	crontab := cron.New()
	_, err := crontab.AddFunc("*/1 * * * *", database.Ping) // 分别代表 分 时 周 月 星期
	if err != nil {
		slog.Error("crontab add func failed", "error", err)
	}
	crontab.Start()
}
```

## Step 10. 监控

### 修改 yaml

```go
scrape_configs:
  - job_name: "prometheus"
    static_configs:
      - targets: ["localhost:9090"]

  - job_name: "my_go_app"
    static_configs:
      - targets: ["localhost:8080"]
```

### `metric`

```go
package handler

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Counter是一个积累量（单调增），跟历史值有关
	requestCounter = promauto.NewCounterVec(prometheus.CounterOpts{Name: "request_counter"}, []string{"service", "interface"}) //此处指定了2个Label
	// Gauge是每个记录是独立的
	requestTimer = promauto.NewGaugeVec(prometheus.GaugeOpts{Name: "request_timer"}, []string{"service", "interface"})
)

// MetricHandler 返回每个接口的调用次数和调用时间
func MetricHandler(ctx *gin.Context) {
	// 记录开始时间
	start := time.Now()

	// 执行后面的中间件
	ctx.Next()

	// 当前接口调用 url, 需要对路径进行处理
	path := mapURL(ctx)

	requestCounter.WithLabelValues("gopostery", path).Inc()                                        // 计数器 + 1 即可
	requestTimer.WithLabelValues("gopostery", path).Set(float64(time.Since(start).Milliseconds())) // 计时器记录从 start 到现在过了多久
}

var mpRestful = map[string]string{"id": "id"}

func mapURL(ctx *gin.Context) string {
	url := ctx.Request.URL.Path
	// ctx.Params 返回请求参数切片, 切片中每个元素为 {Key, Val}
	for _, param := range ctx.Params {
		if value, ok := mpRestful[param.Key]; ok != false {
			url = strings.Replace(url, param.Value, value, 1) // 把具体值换成抽象 eg : /posts/delete/3 -> /posts/delete/:id
		}
	}
	return url
}
```

### Prometheus

![Snipaste_2025-11-27_22-32-06.png](attachment:6f49f792-52ad-498d-90b3-071b82a2c0ff:Snipaste_2025-11-27_22-32-06.png)

![Snipaste_2025-11-27_22-32-59.png](attachment:ea37d1a8-ac85-4396-8e9a-4882c70799d5:Snipaste_2025-11-27_22-32-59.png)

### Grafana

![Snipaste_2025-11-27_22-47-44.png](attachment:3be8d35e-aad4-49c6-8bb9-2d9525045f18:Snipaste_2025-11-27_22-47-44.png)

## Step 11. 改造长短 Token

### `Redis` 连接

```go
package redis

import (
	"log/slog"
	"sync"

	"github.com/go-redis/redis"
	"github.com/yzletter/go-postery/utils"
)

var (
	GoPosteryRedisClient *redis.Client
	redisOnce            sync.Once
)

// ConnectToRedis 连接到 MySQL 数据库, 生成一个 *redis.Client 赋给全局数据库变量 GoPosteryRedisClient
func ConnectToRedis(confDir, confFileName, confFileType string) {
	// 初始化 Viper 进行配置读取
	viper := utils.InitViper(confDir, confFileName, confFileType)
	host := viper.GetString("redis.host")
	port := viper.GetString("redis.port")
	db := viper.GetInt("redis.db")

	redisAddr := host + ":" + port // 拼接地址
	redisOption := &redis.Options{
		Addr: redisAddr,
		DB:   db,
	}

	// 连接到数据库
	redisOnce.Do(func() {
		GoPosteryRedisClient = redis.NewClient(redisOption)
	})

	// 尝试 ping 通
	if err := GoPosteryRedisClient.Ping().Err(); err != nil { // 须加上.Err(), 否则会报 ping 通错
		slog.Error("connect to Redis failed", "error", err)
		panic(err)
	} else {
		slog.Info("connect to Redis succeed")
	}
}

// Ping ping 一下数据库 保持连接
func Ping() {
	if GoPosteryRedisClient != nil {
		err := GoPosteryRedisClient.Ping().Err()
		if err != nil {
			slog.Info("ping GoPosteryRedisClient failed")
			return
		}
		slog.Info("ping GoPosteryRedisClient succeed")
		return
	}
}

func CloseConnection() {
	if GoPosteryRedisClient != nil {
		err := GoPosteryRedisClient.Close()
		if err != nil {
			slog.Info("close GoPosteryRedisClient failed")
			return
		}
		slog.Info("close GoPosteryRedisClient succeed")
		return
	}
}
```

## Step 12. 重构

```go
分层

Repository ——> 操作数据库
Service ——> 进行具体业务逻辑，调用 Repository 操作数据库
Handler ——> 从 http 中获取参数，调用 Service，将结果返回给 http
Middleware ——> 拦截器，可以调用 Service 辅助处理，但不应该承载业务逻辑

Request
   ↓
Middleware (Auth / Logger / Recovery)
   ↓
Handler (解析 HTTP)
   ↓
Service (业务逻辑)
   ↓
Repository (数据库)
   ↓
Response
```

