package infra

import (
	"log/slog"
	"os"
	"sync"

	rmq_client "github.com/apache/rocketmq-clients/golang/v5"
	"github.com/apache/rocketmq-clients/golang/v5/credentials"
	"github.com/yzletter/go-postery/lottery/conf"
	"github.com/yzletter/go-postery/lottery/config"
)

var (
	producer rmq_client.Producer
	consumer rmq_client.SimpleConsumer
	pOnce    sync.Once
	cOnce    sync.Once
)

type RocketMQ struct {
	RocketProducer rmq_client.Producer
	RocketConsumer rmq_client.SimpleConsumer
}

func Init(config config.RocketMQConfig) *RocketMQ {
	// 初始化 RocketMQ 日志
	os.Setenv(rmq_client.CLIENT_LOG_ROOT, "./logs")
	os.Setenv(rmq_client.CLIENT_LOG_FILENAME, "rocketmq.log") // 封装的是 Zap log
	rmq_client.ResetLogger()

	rocketProducer := newProducer(config.Addr)
	rocketConsumer := newConsumer(config.Addr)
	return &RocketMQ{
		RocketProducer: rocketProducer,
		RocketConsumer: rocketConsumer,
	}
}

func newProducer(proxyEndpoint string) rmq_client.Producer {
	pOnce.Do(func() {
		// 初始化过
		if producer != nil {
			return
		}

		// 未初始化过
		var err error
		producer, err = rmq_client.NewProducer(
			&rmq_client.Config{
				Endpoint:      proxyEndpoint,
				NameSpace:     "",
				ConsumerGroup: "",
				Credentials:   &credentials.SessionCredentials{},
			},
			rmq_client.WithTopics(
				conf.RocketLotteryTopic,
			),
		)
		if err != nil {
			slog.Error("初始化 RocketMQ Producer 失败 ...", "error", err)
		}

		err = producer.Start()
		if err != nil {
			slog.Error("启动 RocketMQ Producer 失败 ...", "error", err)
		}
	})

	slog.Info("初始化 RocketMQ Producer 成功 ...")
	return producer
}

func newConsumer(proxyEndpoint string) rmq_client.SimpleConsumer {
	cOnce.Do(func() {
		// 初始化过
		if consumer != nil {
			return
		}

		// 未初始化过
		var err error
		consumer, err = rmq_client.NewSimpleConsumer(
			&rmq_client.Config{
				Endpoint:      proxyEndpoint, // Proxy 地址
				Credentials:   &credentials.SessionCredentials{},
				ConsumerGroup: conf.RocketLotteryConsumerGroup, // 消费方需要指定组
				NameSpace:     "",
			},
			rmq_client.WithSimpleAwaitDuration(conf.RocketAwaitDuration),
			rmq_client.WithSimpleSubscriptionExpressions(
				map[string]*rmq_client.FilterExpression{
					conf.RocketLotteryTopic: rmq_client.SUB_ALL, // 订阅该 Topic 下所有 Tag
				}),
		)
		if err != nil {
			slog.Error("初始化 RocketMQ Consumer 失败 ...", "error", err)
		}

		err = consumer.Start()
		if err != nil {
			slog.Error("启动 RocketMQ Consumer 失败 ...", "error", err)
		}
	})

	slog.Info("初始化 RocketMQ Consumer 成功 ...")
	return consumer
}

func Close() {
	if producer != nil {
		producer.GracefulStop()
	}
	if consumer != nil {
		consumer.GracefulStop()
	}
}
