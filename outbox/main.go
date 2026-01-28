package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/yzletter/go-postery/outbox/conf"
	infraKafka "github.com/yzletter/go-postery/outbox/infra/kafka"
	infraMySQL "github.com/yzletter/go-postery/outbox/infra/mysql"
	infraSlog "github.com/yzletter/go-postery/outbox/infra/slog"
	"github.com/yzletter/go-postery/outbox/infra/viper"
)

func ListenTermSignal(f func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	sig := <-c
	log.Println("receive term signal " + sig.String() + ", going to exit")
	f()
	os.Exit(0)
}

func main() {
	// Infra 层
	infraSlog.InitSlog(conf.LogFilePath)                                   // 初始化 slog
	_ = infraMySQL.Init("./conf", "db", viper.YAML, "./logs")              // 初始化 MySQL
	KafkaProducer := infraKafka.InitProducer([]string{conf.KafkaEndpoint}) // 初始化 Kafka 生产方

	// 开启协程
	ctx, cancel := context.WithCancel(context.Background())
	go infraMySQL.ScanOutbox(ctx, KafkaProducer) // 开启扫表发消息协程

	ListenTermSignal(cancel)
}
