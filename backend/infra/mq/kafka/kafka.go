package kafka

import (
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
)

var (
	producer *kafka.Writer
	ponce    sync.Once
)

func InitProducer(brokers []string) *kafka.Writer {
	ponce.Do(func() {
		producer = &kafka.Writer{
			Addr:                   kafka.TCP(brokers...),
			Balancer:               &kafka.Hash{},
			RequiredAcks:           kafka.RequireAll,
			AllowAutoTopicCreation: true,
			WriteTimeout:           10 * time.Second,
			ReadTimeout:            10 * time.Second,
		}
	})
	return producer
}

func InitBatchProducer(brokers []string) *kafka.Writer {
	ponce.Do(func() {
		producer = &kafka.Writer{
			Addr:                   kafka.TCP(brokers...),
			Balancer:               &kafka.Hash{},
			MaxAttempts:            3,
			BatchSize:              100,
			BatchTimeout:           10 * time.Millisecond,
			ReadTimeout:            10 * time.Second,
			WriteTimeout:           5 * time.Second,
			RequiredAcks:           kafka.RequireAll,
			Async:                  false,
			AllowAutoTopicCreation: true,
		}
	})
	return producer
}

// InitConsumer 初始化 Kafka 消费组，没有已提交位点时从最早的消息开始。
func InitConsumer(brokers []string, groupID string, topics ...string) (*kafka.ConsumerGroup, error) {
	return kafka.NewConsumerGroup(kafka.ConsumerGroupConfig{
		ID:          groupID,
		Brokers:     brokers,
		Topics:      topics,
		StartOffset: kafka.FirstOffset,
	})
}
