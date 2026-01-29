package service

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/markdown"
	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/recursive"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/document"
	eino_model "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
	"github.com/qdrant/go-client/qdrant"
	"github.com/segmentio/kafka-go"
	"github.com/yzletter/go-postery/agent/dto"
	"github.com/yzletter/go-postery/agent/model"
	"github.com/yzletter/go-postery/agent/repository"
	"github.com/yzletter/go-postery/agent/service/ports"
	agent_grpc "github.com/yzletter/go-postery/api/proto/agent/v1"
	post_grpc "github.com/yzletter/go-postery/api/proto/post/v1"
	"github.com/yzletter/go-postery/errno"
	post_conf "github.com/yzletter/go-postery/post/conf"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type agentService struct {
	agentRepo           repository.AgentRepository
	agentKafkaConsumer  *kafka.Reader
	qdrantKafkaConsumer *kafka.Reader
	embedder            ports.Embedder
	llmModel            eino_model.ToolCallingChatModel
	idGenerator         ports.IDGenerator
	agent_grpc.UnimplementedAgentServiceServer
}

func NewAgentService(agentRepo repository.AgentRepository, agentKafkaConsumer *kafka.Reader, qdrantKafkaConsumer *kafka.Reader, embedder ports.Embedder, llmModel eino_model.ToolCallingChatModel, idGenerator ports.IDGenerator) AgentService {
	return &agentService{
		agentRepo:                       agentRepo,
		agentKafkaConsumer:              agentKafkaConsumer,
		qdrantKafkaConsumer:             qdrantKafkaConsumer,
		embedder:                        embedder,
		llmModel:                        llmModel,
		idGenerator:                     idGenerator,
		UnimplementedAgentServiceServer: agent_grpc.UnimplementedAgentServiceServer{},
	}
}

func (svc *agentService) Chat(ctx context.Context, req *agent_grpc.ChatRequest) (*agent_grpc.ChatResponse, error) {
	var empty = new(agent_grpc.ChatResponse)

	// todo 拉取历史记录
	//messages, errs := svc.agentRepo.GetMessagesBySessionID(ctx, sessionID)
	//if errs != nil {
	//	if errors.Is(errs, repository.ErrServerInternal) {
	//		return empty, errno.ErrServerInternal
	//	}
	//}

	//messages := []string{}

	// RAG 召回
	knowledge, err := svc.agentRepo.Retrieve(ctx, req.Query, 0.5, 3) // 召回分数 > 0.5 的三条
	if err != nil {
		return empty, errno.ErrServerInternal
	}

	//// 聚合信息
	//data := map[string]any{
	//	"document":         knowledge,
	//	"query":            query,
	//	"history_messages": messages,
	//}

	//newQuery, _ := sonic.MarshalString(data)

	// 创建 Agent
	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Model:       svc.llmModel,
		Name:        "knowledge_service",
		Description: "知识库助手",
		Instruction: "你是网站的知识库助手，请结合所提供的可靠的论坛文章内容以及历史消息记录以及自己的思考，回答用户的问题",
	})

	// 创建 Runner
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           agent,
		EnableStreaming: false,
		CheckPointStore: nil,
	})
	if err != nil {
		return empty, errno.ErrServerInternal
	}

	// 进行询问
	iterator := runner.Query(ctx, fmt.Sprintf("资料如下：%s\n用户问题：%s", knowledge, req.Query))

	// 返回结果
	var lastMsg adk.Message
	for {
		event, ok := iterator.Next()
		if !ok {
			break
		}
		if event.Err != nil {
			log.Fatal(event.Err)
		}
		msg, err := event.Output.MessageOutput.GetMessage()
		if err != nil {
			log.Fatal(err)
		}
		lastMsg = msg
	}

	if lastMsg == nil {
		return dto.ToChatResponse(nil, req.SessionID, nil), nil
	}

	if lastMsg.Role == schema.Assistant && len(lastMsg.Content) > 0 {
		return dto.ToChatResponse(lastMsg, req.SessionID, knowledge), nil
	}

	return dto.ToChatResponse(nil, req.SessionID, nil), nil
}

// 订阅两种消息：1. 文章新建时，进行切分入库 topic 为 index_document，2. 有向量要入 Qdrant upsert_qdrant
func (svc *agentService) StartChunkDocConsumer(ctx context.Context) {
	backoff := time.Second
	for {
		select {
		case <-ctx.Done():
			slog.Info("关闭 Chunk Doc Consumer 成功 ...")
			return
		default:
			// Fetch 消息
			message, err := svc.agentKafkaConsumer.FetchMessage(ctx)
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
			var payload model.ChunkDocumentEventPayload
			err = sonic.Unmarshal(message.Value, &payload)
			if err != nil {
				slog.Error("invalid message value, skip", "topic", message.Topic, "partition", message.Partition, "offset", message.Offset, "value", string(message.Value), "errs", err)
				// 脏消息 Commit 掉
				_ = svc.agentKafkaConsumer.CommitMessages(ctx, message)
				continue
			}

			// 消费消息
			err = svc.indexDocument(ctx, payload.ID)
			if err != nil {
				slog.Error("Chunk Document Failed", "error", err)
				time.Sleep(time.Second) // 最小退避，避免打爆
				continue                // 不 commit -> 重试
			}

			// 消费成功, 把消息 Commit 掉
			err = svc.agentKafkaConsumer.CommitMessages(ctx, message)
			if err != nil {
				slog.Error("Commit Kafka Message Failed", "id", payload.ID, "topic", message.Topic, "partition", message.Partition, "offset", message.Offset, "errs", err)
				// Commit 失败通常会导致重复消费，但不会丢消息，可接受
				continue
			}
		}
	}
}

func (svc *agentService) StartUpsertQdrantConsumer(ctx context.Context) {
	backoff := time.Second
	for {
		select {
		case <-ctx.Done():
			slog.Info("关闭 Chunk Doc Consumer 成功 ...")
			return
		default:
			// Fetch 消息
			message, err := svc.qdrantKafkaConsumer.FetchMessage(ctx)
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
			var payload model.UpsertQdrantEventPayload
			err = sonic.Unmarshal(message.Value, &payload)
			if err != nil {
				slog.Error("invalid message value, skip", "topic", message.Topic, "partition", message.Partition, "offset", message.Offset, "value", string(message.Value), "errs", err)
				// 脏消息 Commit 掉
				_ = svc.qdrantKafkaConsumer.CommitMessages(ctx, message)
				continue
			}

			// 消费消息
			err = svc.upsertQdrant(ctx, payload.BatchID)
			if err != nil {
				slog.Error("Upsert Vectors Failed", "error", err)
				time.Sleep(time.Second) // 最小退避，避免打爆
				continue                // 不 commit -> 重试
			}

			// 消费成功, 把消息 Commit 掉
			err = svc.qdrantKafkaConsumer.CommitMessages(ctx, message)
			if err != nil {
				slog.Error("Commit Kafka Message Failed", "topic", message.Topic, "partition", message.Partition, "offset", message.Offset, "errs", err)
				// Commit 失败通常会导致重复消费，但不会丢消息，可接受
				continue
			}
		}
	}
}

// 索引文本
func (svc *agentService) indexDocument(ctx context.Context, postID int64) error {
	conn, err := grpc.NewClient(
		"localhost:"+post_conf.Port,
		grpc.WithTransportCredentials(insecure.NewCredentials()), // 设置传输安全
	)

	postClient := post_grpc.NewPostServiceClient(conn)

	// 读文本
	post, err := postClient.GetDetailByID(ctx, &post_grpc.GetDetailByIDRequest{
		PostID:     postID,
		AddViewCnt: false,
	})
	if err != nil {
		slog.Error("获取文本失败", "pid", postID, "error", err)
		return err
	}

	// 转为 Document
	docs := []*schema.Document{{
		ID:       "",
		Content:  post.Content,
		MetaData: nil,
	}}

	// 切分文本
	chunks, err := transform(ctx, docs, int(post.ContentType))
	if err != nil {
		slog.Error("切分文本失败", "pid", postID, "error", err)
		return err
	}

	chunkModels := make([]*model.Chunk, 0, len(chunks))

	// 指定批次
	batchID := svc.idGenerator.NextID()

	for idx, chunk := range chunks {
		chunkID := stableChunkID(post.ID, idx) // 根据 PID + idx 生成固定的 ChunkID, 防止重试生成一套新 ID，导致数据库迅速膨胀

		// 用于入 MySQL
		chunkModels = append(chunkModels, &model.Chunk{
			ID:      chunkID,
			Content: chunk.Content,
			BatchID: batchID,
		})
	}

	var payload model.UpsertQdrantEventPayload
	payload.BatchID = batchID
	value, _ := sonic.MarshalString(payload)
	event := &model.Event{
		ID:           svc.idGenerator.NextID(),
		Topic:        "upsert_qdrant",
		MessageKey:   "upsert_vectors",
		MessageValue: value,
	}

	err = svc.agentRepo.CreateChunksWithOutbox(ctx, chunkModels, event) // 事务写 Chunk 表和 Outbox 表, 异步入 Qdrant 库
	if err != nil {
		slog.Error("MySQL Create Chunk Failed", "error", err)
		return err
	}

	return nil
}

func (svc *agentService) upsertQdrant(ctx context.Context, BatchID int64) error {
	// 根据 BatchID 查找 Chunks
	chunks, err := svc.agentRepo.GetChunksByBatchID(ctx, BatchID)
	if err != nil {
		return err
	}

	// 计算 Embeddings
	// 收集文本，一次性 Embedd
	text := make([]string, 0, len(chunks))
	for _, chunk := range chunks {
		text = append(text, chunk.Content)
	}

	// 计算向量
	vectors, err := svc.embedder.Embedding(ctx, text)
	if err != nil {
		slog.Error("计算向量失败", "error", err)
		return err
	}

	points := make([]*qdrant.PointStruct, 0, len(chunks))

	// 构造 Points
	for idx, chunk := range chunks {
		points = append(points, &qdrant.PointStruct{
			Id:      &qdrant.PointId{PointIdOptions: &qdrant.PointId_Uuid{Uuid: chunk.ID}},
			Vectors: &qdrant.Vectors{VectorsOptions: &qdrant.Vectors_Vector{Vector: &qdrant.Vector{Vector: &qdrant.Vector_Dense{Dense: &qdrant.DenseVector{Data: toFloat32(vectors[idx])}}}}},
		})
	}

	// 插入 Qdrant
	err = svc.agentRepo.UpsertVectorPoints(ctx, points)
	if err != nil {
		return err
	}

	return nil
}

// transform 切分文本
func transform(ctx context.Context, docs []*schema.Document, biz int) ([]*schema.Document, error) {
	// 解析器
	var splitter document.Transformer
	var err error

	if biz == 0 { // 普通文本
		splitter, err = recursive.NewSplitter(ctx, &recursive.Config{
			ChunkSize:   500,                                                  // 必需：目标片段大小
			OverlapSize: 100,                                                  // 可选：片段重叠大小
			Separators:  []string{"\n\n", "\n", ".", "?", "!", "。", "？", "！"}, // 可选：分隔符列表
			LenFunc:     nil,                                                  // 可选：自定义长度计算函数
			KeepType:    recursive.KeepTypeNone,                               // 可选：分隔符保留策略
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
			ChunkSize:   500,                                                  // 必需：目标片段大小
			OverlapSize: 100,                                                  // 可选：片段重叠大小
			Separators:  []string{"\n\n", "\n", ".", "?", "!", "。", "？", "！"}, // 可选：分隔符列表
			LenFunc:     nil,                                                  // 可选：自定义长度计算函数
			KeepType:    recursive.KeepTypeNone,                               // 可选：分隔符保留策略
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

func toFloat32(vector []float64) []float32 {
	rect := make([]float32, len(vector))
	for i, ele := range vector {
		rect[i] = float32(ele)
	}
	return rect
}

// 生成固定 ChunkID
func stableChunkID(postID int64, idx int) string {
	// 你也可以用 sha1/xxhash 做得更短
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(
		fmt.Sprintf("%d:%d", postID, idx),
	)).String()
}
