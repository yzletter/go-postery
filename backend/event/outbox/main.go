package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/yzletter/go-postery/backend/conf"
	infraEtcd "github.com/yzletter/go-postery/backend/infra/cache/etcd"
	infraMySQL "github.com/yzletter/go-postery/backend/infra/database/mysql"
	infraKafka "github.com/yzletter/go-postery/backend/infra/mq/kafka"
	infraSlog "github.com/yzletter/go-postery/backend/infra/slog"
	"github.com/yzletter/go-postery/backend/micro/outbox/service"
)

func ListenTermSignal(f func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	sig := <-c
	log.Println("receive term signal " + sig.String() + ", going to exit")
	f()
	os.Exit(0)
}

const (
	Service   = "outbox_service" // 微服务名
	GoPostery = "go_postery"     // GoPostery 公共配置前缀
)

var (
	prefix       = ""
	EtcdEndPoint string // etcd 地址
)

func main() {
	// 启动参数, 默认线上环境
	env := flag.String("env", "production", "运行环境: local/production")
	flag.Parse()

	// 本地测试
	if *env == "local" {
		prefix = "test_"
		EtcdEndPoint = "localhost:12379"
	} else {
		EtcdEndPoint = "172.16.131.223:2379"
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Remote Config Center
	EtcdClient := infraEtcd.Init([]string{EtcdEndPoint}) // Init Etcd

	// 加载公共配置
	CommonMicroConf := conf.LoadCommonMicroConf(ctx, EtcdClient, prefix+GoPostery+"_")
	fmt.Printf("%s Init Common Config Success %+v\n", prefix+Service, CommonMicroConf)
	// 加载私有配置
	OutboxServiceConf := conf.LoadOutboxServiceConfig(ctx, EtcdClient, prefix+Service+"_")
	fmt.Printf("%s Init OutboxService Config Success %+v\n", prefix+Service, OutboxServiceConf)

	// gRPC Common Infrastructure
	infraSlog.InitSlog(OutboxServiceConf.Log) // Init Slog

	// Infra 层
	MySQLGormDB := infraMySQL.Init(CommonMicroConf.MySQL)                          // 初始化 MySQL
	KafkaProducer := infraKafka.InitProducer([]string{CommonMicroConf.Kafka.Addr}) // 初始化 Kafka 生产方

	// 开启协程
	go service.ScanOutbox(ctx, MySQLGormDB, KafkaProducer) // 开启扫表发消息协程

	ListenTermSignal(func() {
		infraKafka.Close()
		infraMySQL.Close()
		cancel()
	})
}
