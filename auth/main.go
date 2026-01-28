package main

import (
	"log/slog"
	"net"

	auth_grpc "github.com/yzletter/go-postery/api/proto/auth/v1"
	"github.com/yzletter/go-postery/auth/conf"
	conf2 "github.com/yzletter/go-postery/auth/conf"
	infraMySQL "github.com/yzletter/go-postery/auth/infra/mysql"
	infraRedis "github.com/yzletter/go-postery/auth/infra/redis"
	"github.com/yzletter/go-postery/auth/infra/security"
	infraSlog "github.com/yzletter/go-postery/auth/infra/slog"
	"github.com/yzletter/go-postery/auth/infra/snowflake"
	"github.com/yzletter/go-postery/auth/infra/viper"
	"github.com/yzletter/go-postery/auth/repository"
	"github.com/yzletter/go-postery/auth/repository/cache"
	"github.com/yzletter/go-postery/auth/repository/dao"
	"github.com/yzletter/go-postery/auth/service"
	"google.golang.org/grpc"
)

func main() {
	// Infra 层
	infraSlog.InitSlog(conf.LogFilePath)                                 // 初始化 slog
	MySQLGormDB := infraMySQL.Init("./conf", "db", viper.YAML, "./logs") // 初始化 MySQL
	RedisClient := infraRedis.Init("./conf", "cache", viper.YAML)        // 初始化 Redis
	IDGenerator := snowflake.NewSnowflakeIDGenerator(0)                  // 初始化 雪花算法
	PasswordHasher := security.NewBcryptPasswordHasher(0)                // 初始化 密码哈希器
	JwtManager := security.NewJwtManager(conf2.JwtTokenKey)
	// Cache 层
	AuthCache := cache.NewAuthCache(RedisClient)
	// DAO 层
	AuthDAO := dao.NewAuthDAO(MySQLGormDB)
	// Repository 层
	AuthRepo := repository.NewAuthRepository(AuthDAO, AuthCache)
	// Service 层
	AuthService := service.NewAuthService(AuthRepo, JwtManager, PasswordHasher, IDGenerator) // 注册 AuthService

	// 监听本地端口
	lis, err := net.Listen("tcp", "localhost:"+conf.Port)
	if err != nil {
		panic(err)
	}

	// 启动 grpc 服务
	server := grpc.NewServer()
	auth_grpc.RegisterAuthServiceServer(server, AuthService) // 注册服务
	if err := server.Serve(lis); err != nil {
		slog.Error("Auth grpc Server Start Failed", "error", err)
		panic(err)
	}
}
