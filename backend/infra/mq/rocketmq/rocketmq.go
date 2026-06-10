package infra

import (
	"log/slog"
	"os"
	"sync"
	"time"

	rmq_client "github.com/apache/rocketmq-clients/golang/v5"
	"github.com/apache/rocketmq-clients/golang/v5/credentials"
	"github.com/yzletter/go-postery/backend/conf"
)

var (
	producer rmq_client.Producer
	consumer rmq_client.SimpleConsumer
	pOnce    sync.Once
	cOnce    sync.Once
	pFlag    bool
	cFlag    bool
)

type RocketMQ struct {
	RocketProducer rmq_client.Producer
	RocketConsumer rmq_client.SimpleConsumer
}

func Init(config conf.RocketMQConfig, topic string, group string, duration time.Duration) *RocketMQ {
	// 初始化 RocketMQ 日志
	_ = os.Setenv(rmq_client.CLIENT_LOG_ROOT, "./logs")
	_ = os.Setenv(rmq_client.CLIENT_LOG_FILENAME, "rocketmq.log") // 封装的是 Zap log
	rmq_client.ResetLogger()

	rocketProducer := newProducer(config.Addr, topic)
	rocketConsumer := newConsumer(config.Addr, topic, group, duration)
	return &RocketMQ{
		RocketProducer: rocketProducer,
		RocketConsumer: rocketConsumer,
	}
}

func newProducer(proxyEndpoint string, topic string) rmq_client.Producer {
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
				topic,
			),
		)
		if err != nil {
			slog.Error("初始化 RocketMQ Producer 失败 ...", "error", err)
			return
		}

		if err = producer.Start(); err != nil {
			slog.Error("启动 RocketMQ Producer 失败 ...", "error", err)
			return
		}

		pFlag = true
	})

	if pFlag {
		slog.Info("初始化 RocketMQ Producer 成功 ...")
	}
	return producer
}

func newConsumer(proxyEndpoint string, topic string, group string, duration time.Duration) rmq_client.SimpleConsumer {
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
				ConsumerGroup: group, // 消费方需要指定组
				NameSpace:     "",
			},
			rmq_client.WithSimpleAwaitDuration(duration),
			rmq_client.WithSimpleSubscriptionExpressions(
				map[string]*rmq_client.FilterExpression{
					topic: rmq_client.SUB_ALL, // 订阅该 Topic 下所有 Tag
				}),
		)
		if err != nil {
			slog.Error("初始化 RocketMQ Consumer 失败 ...", "error", err)
			return
		}

		if err = consumer.Start(); err != nil {
			slog.Error("启动 RocketMQ Consumer 失败 ...", "error", err)
			return
		}

		cFlag = true
	})

	if cFlag {
		slog.Info("初始化 RocketMQ Consumer 成功 ...")
	}
	return consumer
}

func Close() {
	if producer != nil {
		_ = producer.GracefulStop()
	}
	if consumer != nil {
		_ = consumer.GracefulStop()
	}
}
