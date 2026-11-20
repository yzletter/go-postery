package service

import (
	"context"
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/protocol/metadata"
	"github.com/segmentio/kafka-go/protocol/produce"
	"github.com/yzletter/go-postery/backend/micro/outbox/model"
)

func TestSendOutboxEventsIsolatesTopicErrors(t *testing.T) {
	produced := make(chan struct{})
	writer := &kafka.Writer{
		Addr: kafka.TCP("localhost:9092"),
		Balancer: kafka.BalancerFunc(func(msg kafka.Message, partitions ...int) int {
			return partitions[msg.Key[0]-'0']
		}),
		BatchTimeout: time.Millisecond,
		MaxAttempts:  1,
		RequiredAcks: kafka.RequireAll,
		Transport: outboxTestTransport(func(ctx context.Context, req kafka.Request) (kafka.Response, error) {
			switch req := req.(type) {
			case *metadata.Request:
				topic := metadata.ResponseTopic{Name: req.TopicNames[0]}
				if topic.Name == "denied" {
					// 等正常 Topic 开始写入，验证元数据请求卡住时其他组仍能发送。
					select {
					case <-produced:
					case <-ctx.Done():
						return nil, ctx.Err()
					}
					topic.ErrorCode = int16(kafka.TopicAuthorizationFailed)
				} else {
					topic.Partitions = []metadata.ResponsePartition{{PartitionIndex: 0}, {PartitionIndex: 1}}
				}
				return &metadata.Response{Topics: []metadata.ResponseTopic{topic}}, nil
			case *produce.Request:
				topic := req.Topics[0]
				partition := produce.ResponsePartition{Partition: topic.Partitions[0].Partition}
				if partition.Partition == 0 {
					close(produced)
				} else {
					partition.ErrorCode = int16(kafka.MessageSizeTooLarge)
				}
				return &produce.Response{Topics: []produce.ResponseTopic{{
					Topic: topic.Topic, Partitions: []produce.ResponsePartition{partition},
				}}}, nil
			default:
				return nil, fmt.Errorf("unexpected Kafka request %T", req)
			}
		}),
	}
	t.Cleanup(func() { _ = writer.Close() })
	ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
	defer cancel()

	got := sendOutboxEvents(ctx, writer, []*model.OutboxEvent{
		{Topic: "denied", MessageKey: "0"},
		{Topic: "posts", MessageKey: "0"},
		{Topic: "denied", MessageKey: "0"},
		{Topic: "posts", MessageKey: "1"},
	})
	want := []error{kafka.TopicAuthorizationFailed, nil, kafka.TopicAuthorizationFailed, kafka.MessageSizeTooLarge}
	if len(got) != len(want) {
		t.Fatalf("got %d send results, want %d", len(got), len(want))
	}
	for i, err := range got {
		if !errors.Is(err, want[i]) {
			t.Errorf("result for event %d = %v, want %v", i, err, want[i])
		}
	}
}

type outboxTestTransport func(context.Context, kafka.Request) (kafka.Response, error)

func (f outboxTestTransport) RoundTrip(ctx context.Context, _ net.Addr, req kafka.Request) (kafka.Response, error) {
	return f(ctx, req)
}
