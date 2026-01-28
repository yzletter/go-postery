package main

import (
	"log/slog"
	"net"

	post_grpc "github.com/yzletter/go-postery/api/proto/post/v1"
	"github.com/yzletter/go-postery/post/conf"
	infraMySQL "github.com/yzletter/go-postery/post/infra/mysql"
	infraRedis "github.com/yzletter/go-postery/post/infra/redis"
	infraSlog "github.com/yzletter/go-postery/post/infra/slog"
	"github.com/yzletter/go-postery/post/infra/snowflake"
	"github.com/yzletter/go-postery/post/infra/viper"
	"github.com/yzletter/go-postery/post/repository"
	"github.com/yzletter/go-postery/post/repository/cache"
	"github.com/yzletter/go-postery/post/repository/dao"
	"github.com/yzletter/go-postery/post/service"
	"google.golang.org/grpc"
)

func main() {
	// Infra 层
	infraSlog.InitSlog(conf.LogFilePath)                                 // 初始化 slog
	RedisClient := infraRedis.Init("./conf", "cache", viper.YAML)        // 初始化 Redis
	MySQLGormDB := infraMySQL.Init("./conf", "db", viper.YAML, "./logs") // 初始化 MySQL
	IDGenerator := snowflake.NewSnowflakeIDGenerator(0)                  // 初始化 雪花算法
	// Cache 层
	PostCache := cache.NewPostCache(RedisClient)
	// DAO 层
	PostDAO := dao.NewPostDAO(MySQLGormDB)
	LikeDAO := dao.NewLikeDAO(MySQLGormDB)
	TagDAO := dao.NewTagDAO(MySQLGormDB)
	CommentDAO := dao.NewCommentDAO(MySQLGormDB)
	// Repository 层
	PostRepo := repository.NewPostRepository(PostDAO, PostCache) // 注册 PostRepository
	LikeRepo := repository.NewLikeRepository(LikeDAO)            // 注册 LikeRepository
	TagRepo := repository.NewTagRepository(TagDAO)               // 注册 TagRepository
	CommentRepo := repository.NewCommentRepository(CommentDAO)   // 注册 CommentRepository
	// Service 层
	PostService := service.NewPostService(PostRepo, LikeRepo, TagRepo, CommentRepo, IDGenerator) // 注册 postSvc

	// 监听本地端口
	lis, err := net.Listen("tcp", "localhost:"+conf.Port)
	if err != nil {
		panic(err)
	}

	// 启动 grpc 服务
	server := grpc.NewServer()
	post_grpc.RegisterPostServiceServer(server, PostService) // 注册服务
	if err := server.Serve(lis); err != nil {
		slog.Error("Post grpc Server Start Failed", "error", err)
		panic(err)
	}
}
