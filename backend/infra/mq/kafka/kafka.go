package kafka

import (
	"log/slog"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/yzletter/go-postery/backend/conf"
)

var (
	producer  *kafka.Writer
	consumers []*kafka.Reader
	ponce     sync.Once
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

func InitConsumer(config conf.KafkaConfig, topic string, groupID string) *kafka.Reader {
	consumer := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{config.Addr},
		Topic:          topic,
		GroupID:        groupID,
		CommitInterval: 0,
		StartOffset:    kafka.FirstOffset,
	})
	consumers = append(consumers, consumer)
	return consumer
}

func Close() {
	if producer != nil {
		if err := producer.Close(); err != nil {
			slog.Error("Close Kafka Writer Failed", "error", err)
		} else {
			slog.Info("Close Kafka Writer Success")
		}
	}

	for _, consumer := range consumers {
		if err := consumer.Close(); err != nil {
			slog.Error("Close Kafka Reader Failed", "error", err)
		} else {
			slog.Info("Close Kafka Reader Success")
		}
	}
}
