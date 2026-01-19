package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/markdown"
	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/recursive"
	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/schema"
	"github.com/segmentio/kafka-go"
	"github.com/yzletter/go-postery/model"
	"github.com/yzletter/go-postery/repository"
	"github.com/yzletter/go-postery/service/ports"
)

type agentService struct {
	agentRepo     repository.AgentRepository
	postRepo      repository.PostRepository
	commentRepo   repository.CommentRepository
	kafkaConsumer *kafka.Reader
	embedder      ports.Embedder
	idGenerator   ports.IDGenerator
}

func NewAgentService(agentRepo repository.AgentRepository, postRepo repository.PostRepository, commentRepo repository.CommentRepository, kafkaConsumer *kafka.Reader, embedder ports.Embedder, idGenerator ports.IDGenerator) AgentService {
	return &agentService{
		agentRepo:     agentRepo,
		postRepo:      postRepo,
		commentRepo:   commentRepo,
		kafkaConsumer: kafkaConsumer,
		embedder:      embedder,
		idGenerator:   idGenerator,
	}
}

func (svc *agentService) StartChunkDocConsumer(ctx context.Context) {
	backoff := time.Second
	for {
		select {
		case <-ctx.Done():
			slog.Info("关闭 Chunk Doc Consumer 成功 ...")
			return
		default:
			// Fetch 消息
			message, err := svc.kafkaConsumer.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil { // 正常退出
					return
				}

				slog.Error("Fetch Message From Kafka Failed", "Kafka", "ChunkKafka", "error", err)

				// 简单退避，避免狂刷日志
				time.Sleep(backoff)
				if backoff < 10*time.Second {
					backoff *= 2
				}
				continue
			}

			backoff = time.Second // 重置

			// 解析 JSON
			var payload model.ChunkDocumentEvent
			err = sonic.Unmarshal(message.Value, &payload)
			if err != nil {
				slog.Error("invalid message value, skip", "topic", message.Topic, "partition", message.Partition, "offset", message.Offset, "value", string(message.Value), "err", err)
				// 脏消息 Commit 掉
				_ = svc.kafkaConsumer.CommitMessages(ctx, message)
			}

			// 消费消息
			err = svc.IndexDocument(payload.ID)
			if err != nil {
				slog.Error("Chunk Document Failed", "error", err)
				time.Sleep(time.Second) // 最小退避，避免打爆
				continue                // 不 commit -> 重试
			}

			// 消费成功, 把消息 Commit 掉
			err = svc.kafkaConsumer.CommitMessages(ctx, message)
			if err != nil {
				slog.Error("Commit Kafka Message Failed", "id", payload.ID, "topic", message.Topic, "partition", message.Partition, "offset", message.Offset, "err", err)
				// Commit 失败通常会导致重复消费，但不会丢消息，可接受
				continue
			}
		}
	}
}

// IndexDocument 索引文本
func (svc *agentService) IndexDocument(id int64) error {

	return nil
}

// RetrieveDocument 召回
func (svc *agentService) RetrieveDocument(ctx context.Context, query string, limit int) []*model.Chunk {

}

// Transform 切分文本
func Transform(ctx context.Context, docs []*schema.Document, biz int) ([]*schema.Document, error) {
	// 解析器
	var splitter document.Transformer
	var err error

	if biz == 0 { // 普通文本
		splitter, err = recursive.NewSplitter(ctx, &recursive.Config{
			ChunkSize:   500,                                   // 必需：目标片段大小
			OverlapSize: 100,                                   // 可选：片段重叠大小
			Separators:  []string{"\n", "\n\n", ".", "?", "!"}, // 可选：分隔符列表
			LenFunc:     nil,                                   // 可选：自定义长度计算函数
			KeepType:    recursive.KeepTypeNone,                // 可选：分隔符保留策略
		})
		if err != nil {
			return []*schema.Document{}, err
		}
	} else { // MarkDown 文本
		// 先切 header
		headerSplitter, err := markdown.NewHeaderSplitter(ctx, &markdown.HeaderConfig{
			Headers: map[string]string{
				"#":  "",
				"##": "",
			},
		})
		if err != nil {
			return []*schema.Document{}, err
		}

		headerTransformedDoc, err := headerSplitter.Transform(ctx, docs)
		if err != nil {
			return []*schema.Document{}, err
		}
		docs = headerTransformedDoc

		// 再递归切
		splitter, err = recursive.NewSplitter(ctx, &recursive.Config{
			ChunkSize:   500,                                   // 必需：目标片段大小
			OverlapSize: 100,                                   // 可选：片段重叠大小
			Separators:  []string{"\n", "\n\n", ".", "?", "!"}, // 可选：分隔符列表
			LenFunc:     nil,                                   // 可选：自定义长度计算函数
			KeepType:    recursive.KeepTypeNone,                // 可选：分隔符保留策略
		})
		if err != nil {
			return []*schema.Document{}, err
		}
	}
	
	// 进行切分
	transformedDocs, err := splitter.Transform(ctx, docs)
	if err != nil {
		return []*schema.Document{}, err
	}

	// todo metadata
	for _, doc := range transformedDocs {
		doc.MetaData = map[string]any{}
	}

	return transformedDocs, nil
}
