package kafka

import (
	"log/slog"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/yzletter/go-postery/microservice-backend/user/conf"
	"github.com/yzletter/go-postery/microservice-backend/user/config"
)

var (
	producer  *kafka.Writer
	consumers []*kafka.Reader
	ponce     sync.Once
)

func InitProducer(brokers []string) *kafka.Writer {
	ponce.Do(func() {
		producer = &kafka.Writer{
			Addr:                   kafka.TCP(brokers...), // 不定长参数，支持传入多个broker的ip:port
			Balancer:               &kafka.Hash{},         // 把message的key进行hash，确定partition
			RequiredAcks:           kafka.RequireAll,      // RequireNone不需要等待ack返回，效率最高，安全性最低；RequireOne只需要确保Leader写入成功就可以发送下一条消息；RequiredAcks需要确保Leader和所有Follower都写入成功才可以发送下一条消息。
			AllowAutoTopicCreation: true,                  // Topic不存在时自动创建。生产环境中一般设为false，由运维管理员创建Topic并配置partition数目
			WriteTimeout:           10 * time.Second,      // 设定写超时
			ReadTimeout:            10 * time.Second,      // 建议也补上
		}
	})
	return producer
}

func InitConsumer(config config.KafkaConfig) *kafka.Reader {
	consumer := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{config.Addr}, // 支持传入多个broker的ip:port
		Topic:          conf.KafkaTopic,
		GroupID:        conf.KafkaGroup,   // 一个Group内消费到的消息不会重复。注意：如果不指定GroupID，则只能消费到1个partition里的数据，所以consumer的个数需要多于partition数据才能把数据消费全
		CommitInterval: 0,                 // 每隔多长时间自动commit一次offset。即一边读一边向kafka上报读到了哪个位置
		StartOffset:    kafka.FirstOffset, // 当一个特定的partition没有commited offset时(比如第一次读一个partition，之前没有commit过)，通过StartOffset指定从第一个还是最后一个位置开始消费。StartOffset的取值要么是FirstOffset要么是LastOffset，LastOffset表示Consumer启动之前生成的老数据不管了。仅当指定了GroupID时，StartOffset才生效。
		// Partition:      0,                 // Partition和GroupID不能同时指定
	})
	consumers = append(consumers, consumer)
	return consumer
}

// Close 关闭 Kafka 连接
func Close() {
	if producer != nil {
		if err := producer.Close(); err != nil {
			slog.Error("关闭 Kafka Writer 失败 ...")
		} else {
			slog.Info("关闭 Kafka Writer 成功 ...")
		}
	}

	for _, consumer := range consumers {
		if err := consumer.Close(); err != nil {
			slog.Error("关闭 Kafka Reader 失败 ...")
		} else {
			slog.Info("关闭 Kafka Reader 成功 ...")
		}
	}
}
