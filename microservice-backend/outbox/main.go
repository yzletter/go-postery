package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	conf2 "github.com/yzletter/go-postery/microservice-backend/outbox/conf"
	"github.com/yzletter/go-postery/microservice-backend/outbox/config"
	infraEtcd "github.com/yzletter/go-postery/microservice-backend/outbox/infra/etcd"
	infraKafka "github.com/yzletter/go-postery/microservice-backend/outbox/infra/kafka"
	infraMySQL "github.com/yzletter/go-postery/microservice-backend/outbox/infra/mysql"
	infraSlog "github.com/yzletter/go-postery/microservice-backend/outbox/infra/slog"
)

func ListenTermSignal(f func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	sig := <-c
	log.Println("receive term signal " + sig.String() + ", going to exit")
	f()
	os.Exit(0)
}

var (
	ServiceName  string // 微服务名
	GoPostery    string // GoPostery 公共配置前缀
	EtcdEndPoint string // etcd 地址
)

func main() {
	// 启动参数, 默认线上环境
	env := flag.String("env", "production", "运行环境: local/production")
	flag.Parse()

	// 本地测试
	if *env == "local" {
		ServiceName = "test_outbox_service"
		GoPostery = "test_go_postery"
		EtcdEndPoint = "localhost:12379"
	} else {
		ServiceName = "outbox_service"
		GoPostery = "go_postery"
		EtcdEndPoint = "172.16.131.223:2379"
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Remote Config Center
	EtcdClient := infraEtcd.Init([]string{EtcdEndPoint})                               // Init Etcd
	Config := config.LoadGlobalConfig(ctx, EtcdClient, ServiceName+"_", GoPostery+"_") // Get Config From Remote Config Center
	fmt.Printf("%s Init Config Success %+v\n", ServiceName, Config)

	// gRPC Common Infrastructure
	infraSlog.InitSlog(Config.Log) // Init Slog

	// Infra 层
	_ = infraMySQL.Init(Config.MySQL)                                       // 初始化 MySQL
	KafkaProducer := infraKafka.InitProducer([]string{conf2.KafkaEndpoint}) // 初始化 Kafka 生产方

	// 开启协程
	go infraMySQL.ScanOutbox(ctx, KafkaProducer) // 开启扫表发消息协程

	ListenTermSignal(cancel)
}
