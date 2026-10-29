package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/yzletter/go-postery/backend/conf"
	infraEtcd "github.com/yzletter/go-postery/backend/infra/cache/etcd"
	infraMySQL "github.com/yzletter/go-postery/backend/infra/database/mysql"
	infraKafka "github.com/yzletter/go-postery/backend/infra/mq/kafka"
	infraSlog "github.com/yzletter/go-postery/backend/infra/slog"
	"github.com/yzletter/go-postery/backend/micro/outbox/config"
	"github.com/yzletter/go-postery/backend/micro/outbox/service"
	"github.com/yzletter/go-postery/backend/utils"
)

const (
	Service       = "outbox_service" // 微服务名
	ServicePrefix = Service + "/"    // 微服务配置前缀

)

func main() {
	// 启动时指定远程配置中心地址
	etcdEndpoint := flag.String("etcd", "localhost:2379", "远程配置中心地址，默认为本地")
	flag.Parse()

	// 初始化根 Context
	ctx, cancel := context.WithCancel(context.Background())

	// 初始化远程配置中心
	EtcdClient := infraEtcd.Init([]string{*etcdEndpoint}) // Init Etcd

	// 加载公共配置
	CommonConf := conf.LoadCommonConf(ctx, EtcdClient, conf.CommonConfPrefix)
	fmt.Printf("%s Init Common Config Success %+v\n", Service, CommonConf)

	// 加载私有配置
	OutboxServiceConf := config.LoadOutboxServiceConfig(ctx, EtcdClient, conf.ServiceConfPrefix+ServicePrefix)
	fmt.Printf("%s Init OutboxService Config Success %+v\n", Service, OutboxServiceConf)

	// 初始化依赖
	infraSlog.InitSlog(OutboxServiceConf.Log) // 初始化 Slog

	// Infrastructure 层
	MySQLGormDB := infraMySQL.Init(CommonConf.MySQL)                          // 初始化 MySQL
	KafkaProducer := infraKafka.InitProducer([]string{CommonConf.Kafka.Addr}) // 初始化 Kafka 生产方

	// 开启协程
	go service.ScanOutbox(ctx, MySQLGormDB, KafkaProducer) // 开启扫表发消息协程

	// 优雅退出
	utils.ListenTermSignal(func() {
		infraKafka.Close()
		infraMySQL.Close()
		KafkaProducer.Close()
		cancel()
	})
}
