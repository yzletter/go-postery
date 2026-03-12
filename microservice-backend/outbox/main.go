package main

import (
	"context"
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

const ServiceName = "outbox_service"

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	// Remote Config Center
	EtcdClient := infraEtcd.Init([]string{"172.16.131.223:2379"})       // Init Etcd
	Config := config.LoadGlobalConfig(ctx, EtcdClient, ServiceName+"_") // Get Config From Remote Config Center
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
