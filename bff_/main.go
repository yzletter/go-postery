package main

import (
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	infraQdarant "github.com/yzletter/go-postery/agent/infra/qdrant"
	infraRedis "github.com/yzletter/go-postery/bff/infra/redis"
	"github.com/yzletter/go-postery/bff/infra/viper"
	"github.com/yzletter/go-postery/bff_/conf"
	"github.com/yzletter/go-postery/bff_/handler"
	"github.com/yzletter/go-postery/bff_/infra/crontab"
	"github.com/yzletter/go-postery/bff_/infra/graceful_stop"
	infraRabbitMQ "github.com/yzletter/go-postery/bff_/infra/rabbitmq"
	"github.com/yzletter/go-postery/bff_/infra/slog"
	"github.com/yzletter/go-postery/bff_/middleware"
	"github.com/yzletter/go-postery/bff_/service"
	infraRocketMQ "github.com/yzletter/go-postery/lottery/infra/rocketmq"
	infraKafka "github.com/yzletter/go-postery/outbox/infra/kafka"
	"github.com/yzletter/go-postery/outbox/infra/mysql"
)

func main() {
	// Infra 层
	RabbitMQ := infraRabbitMQ.Init("./conf", "mq", viper.YAML) // 初始化 RabbitMQ

	// GRPC Service 层
	SearchGRPCSvc := service.NewSearchService("localhost:" + conf.SearchPort)
	PostGRPCSvc := service.NewPostService("localhost:" + conf.PostPort)
	LotteryGRPCSvc := service.NewLotteryService("localhost:" + conf.LotteryPort)
	AgentGRPCSvc := service.NewAgentService("localhost:" + conf.AgentPort)
	SessionGRPCSvc := service.NewSessionService("localhost:" + conf.SessionPort)

	// Service 层
	WebsocketSvc := service.NewWebsocketService(SessionGRPCSvc, RabbitMQ) // 注册 WebsocketService

	// Handler 层

	PostHdl := handler.NewPostHandler(PostGRPCSvc, UserGRPCSvc) // 注册 PostHandler
	SessionHdl := handler.NewSessionHandler(SessionGRPCSvc)     // 注册 SessionHandler
	LotteryHdl := handler.NewLotteryHandler(LotteryGRPCSvc)     // 注册 LotteryHandler
	AgentHdl := handler.NewAgentHandler(AgentGRPCSvc)           // 注册 AgentHandler
	SearchHdl := handler.NewSearchHandler(SearchGRPCSvc)        // 注册 SearchHandler
	WebsocketHdl := handler.NewWebsocketHandler(WebsocketSvc)   // 注册 WebsocketHandler

	// 中间件层

	// 初始化 Crontab
	crontab.NewCrontabBuilder().
		AddFuncWithSpec("*/10 * * * *", infra.Ping).
		AddFuncWithSpec("*/10 * * * *", infraRedis.Ping).
		Build()

	// 初始化 GracefulStop
	graceful_stop.NewGracefulStopBuilder().
		NotifySignal(syscall.SIGINT).NotifySignal(syscall.SIGTERM).
		AddFunc(infraRabbitMQ.Close).AddFunc(infraRocketMQ.Close).AddFunc(infraKafka.Close). // 关消息队列
		AddFunc(infra.Close).AddFunc(infraRedis.Close).AddFunc(infraQdarant.Close).          // 关数据库
		Build()

	// 帖子模块
	posts := v1.Group("/posts")
	{
		posts.GET("", PostHdl.List)                             // POST /api/v1/posts?pageNo=1&pageSize=10				按页获取帖子列表
		posts.GET("/top", PostHdl.Top)                          // GET /api/v1/posts/top								获取热门帖子榜单
		posts.GET("/tags", PostHdl.ListByTagAndPage)            // POST /api/v1/posts/tags?pageNo=1&pageSize=10&tag=go 根据标签按页获取帖子列表
		posts.GET("/:id", PostHdl.Detail)                       // GET /api/v1/posts/:id								获取帖子详情
		posts.GET("/:id/comments", CommentHdl.ListByPage)       // GET /api/v1/posts/:id/comments?pageNo=1&pageSize=10	按页获取帖子评论
		posts.GET("/:id/comments/:cid", CommentHdl.ListReplies) // GET /api/v1/posts/:pid/comments/:cid?pageNo=1&pageSize=10	按页获取主评论回复

		//todo
		authedPosts := posts.Group("")
		authedPosts.Use(AuthRequiredMdl)
		authedPosts.POST("", PostHdl.Create)       // POST /api/v1/posts 		创建帖子
		authedPosts.POST("/:id", PostHdl.Update)   // POST /api/v1/posts/:id 	更新帖子
		authedPosts.DELETE("/:id", PostHdl.Delete) // DELETE /api/v1/posts/:id 	删除帖子

		authedPosts.POST("/:id/comments", CommentHdl.Create)        // POST /api/v1/posts/:id/comments 创建评论
		authedPosts.DELETE("/:id/comments/:cid", CommentHdl.Delete) // DELETE /api/v1/posts/:id/comments/:cid 删除评论
		authedPosts.GET("/:id/likes", PostHdl.IfLike)               // GET /api/v1/posts/:id/likes	查询是否点赞了帖子
		authedPosts.POST("/:id/likes", PostHdl.Like)                // POST /api/v1/posts/:id/likes	点赞帖子
		authedPosts.DELETE("/:id/likes", PostHdl.Unlike)            // DELETE /api/v1/posts/:id/likes 取消点赞帖子
	}

	// 私信模块
	sessions := v1.Group("/sessions")
	sessions.Use(AuthRequiredMdl)
	{
		sessions.GET("", SessionHdl.List)          // GET /api/v1/sessions							获取当前登录用户会话列表
		sessions.DELETE("/:id", SessionHdl.Delete) // DELETE /api/v1/sessions/:id						删除当前会话
	}

	// 即时聊天模块
	im := v1.Group("/ws")
	im.Use(AuthRequiredMdl)
	{
		im.GET("", WebsocketHdl.Connect) // GET /api/v1/ws
	}

	// 抽奖模块
	v1.GET("/gifts", LotteryHdl.GetAllGifts) // GET /api/v1/gifts 获取所有奖品信息
	lottery := v1.Group("/lottery")
	lottery.Use(AuthRequiredMdl)
	{
		lottery.GET("/lucky", LotteryHdl.Lottery)  // GET /api/v1/lottery/lucky 抽奖
		lottery.POST("/giveup", LotteryHdl.GiveUp) // POST /api/v1/lottery/giveup 放弃
		lottery.POST("/pay", LotteryHdl.Pay)       // POST /api/v1/lottery/pay 支付
		lottery.GET("/result", LotteryHdl.Result)  // GET /api/v1/lottery/result 查询结果
	}

	agent := v1.Group("/agent")
	agent.Use(AuthRequiredMdl)
	{
		agent.POST("/chat", AgentHdl.Chat) // POST /api/v1/agent/chat
	}

	search := v1.Group("/search")
	search.Use(AuthRequiredMdl)
	{
		search.POST("", SearchHdl.Search)
	}

	if err := engine.Run("localhost:8765"); err != nil {
		panic(err)
	}
}
