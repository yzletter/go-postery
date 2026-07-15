package rag

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/bytedance/sonic"
	MilvusIndexer "github.com/cloudwego/eino-ext/components/indexer/milvus"
	MilvusRetriever "github.com/cloudwego/eino-ext/components/retriever/milvus"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
	milvusSDK "github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/yzletter/go-postery/backend/micro/interview/model"
	"github.com/yzletter/go-postery/backend/micro/interview/repository/dao"
)

var (
	ErrServerInternal = errors.New("server internal error")
)

type MilvusStore struct {
	client    milvusSDK.Client
	embedder  embedding.Embedder
	retriever *MilvusRetriever.Retriever
}

const (
	CollectionName  = "interview_questions"
	VectorDimension = 1024
)

// 将 *schema.Document 转为 Milvus 的 model.InterviewQuestion 映射
func converter(ctx context.Context, docs []*schema.Document, vectors [][]float64) ([]interface{}, error) {
	questions := make([]interface{}, 0, len(docs))

	for idx, doc := range docs {

		metadata, err := sonic.Marshal(doc.MetaData)
		if err != nil {
			slog.Error("sonic marshal failed", "error", err)
			return nil, dao.ErrServerInternal
		}

		// 将当前 doc 的 []float64 向量转为 []float32
		vec32 := make([]float32, len(vectors[idx]))
		for i, v := range vectors[idx] {
			vec32[i] = float32(v)
		}

		// Eino 的 Document.ID 固定是 string, 写入 Milvus 前转回 int64
		id, err := parseDocumentID(doc.ID)
		if err != nil {
			slog.Error("parse document id failed", "id", doc.ID, "error", err)
			return nil, dao.ErrParamsInvalid
		}

		// 从 MetaData 中获取 UserID 和源文件
		userID, ok := doc.MetaData["user_id"].(int64)
		if !ok {
			slog.Error("parse document user_id failed", "id", doc.ID, "user_id", doc.MetaData["user_id"])
			return nil, dao.ErrParamsInvalid
		}
		sourceFile, _ := doc.MetaData["source_file"].(string)

		// 放入 slice
		questions = append(questions, &model.InterviewQuestion{
			ID:         id,
			Content:    doc.Content,
			Vector:     vec32,
			Metadata:   metadata,
			UserID:     userID,
			SourceFile: sourceFile,
		})
	}

	return questions, nil
}

func vectorConverter(ctx context.Context, vectors [][]float64) ([]entity.Vector, error) {
	res := make([]entity.Vector, 0, len(vectors))
	for _, vector := range vectors {
		fv := make(entity.FloatVector, len(vector))

		for i := range vector {
			fv[i] = float32(vector[i])
		}
		res = append(res, fv)
	}
	return res, nil
}

// parseDocumentID 将 Eino 文档 ID 转回业务侧 int64 ID
func parseDocumentID(id string) (int64, error) {
	if id == "" {
		return 0, fmt.Errorf("empty document id")
	}
	return strconv.ParseInt(id, 10, 64)
}

// milvusDocumentConverter 将 Milvus 查询结果转回 Eino Document
func milvusDocumentConverter(ctx context.Context, result milvusSDK.SearchResult) ([]*schema.Document, error) {
	docs := make([]*schema.Document, result.IDs.Len())
	for i := range docs {
		id, err := result.IDs.GetAsInt64(i)
		if err != nil {
			return nil, fmt.Errorf("failed to get id: %w", err)
		}
		docs[i] = &schema.Document{
			// Document.ID 是第三方 string 边界, 只在这里做格式化
			ID:       strconv.FormatInt(id, 10),
			MetaData: make(map[string]any),
		}
	}

	for _, field := range result.Fields {
		switch field.Name() {
		case "id":
			continue
		case "content":
			for i, doc := range docs {
				content, err := field.GetAsString(i)
				if err != nil {
					return nil, fmt.Errorf("failed to get content: %w", err)
				}
				doc.Content = content
			}
		case "metadata":
			for i, doc := range docs {
				v, err := field.Get(i)
				if err != nil {
					return nil, fmt.Errorf("failed to get metadata: %w", err)
				}
				bytes, ok := v.([]byte)
				if !ok {
					return nil, fmt.Errorf("metadata field has invalid type %T", v)
				}
				if err := sonic.Unmarshal(bytes, &doc.MetaData); err != nil {
					return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
				}
			}
		default:
			for i, doc := range docs {
				v, err := field.Get(i)
				if err != nil {
					return nil, fmt.Errorf("failed to get field %q: %w", field.Name(), err)
				}
				doc.MetaData[field.Name()] = v
			}
		}
	}
	return docs, nil
}

func NewMilvusRAGStore(client milvusSDK.Client, embedder embedding.Embedder, topK int) *MilvusStore {
	initCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 初始化 Indexer, 借用 eino-ext 创建 collection 和索引
	if _, err := MilvusIndexer.NewIndexer(initCtx, &MilvusIndexer.IndexerConfig{
		Client:     client,
		Collection: CollectionName,
		Fields: []*entity.Field{
			entity.NewField().WithName("id").WithDataType(entity.FieldTypeInt64).WithIsPrimaryKey(true),
			entity.NewField().WithName("content").WithDataType(entity.FieldTypeVarChar).WithMaxLength(8192),
			entity.NewField().WithName("vector").WithDataType(entity.FieldTypeFloatVector).WithDim(VectorDimension),
			entity.NewField().WithName("metadata").WithDataType(entity.FieldTypeJSON),
			entity.NewField().WithName("user_id").WithDataType(entity.FieldTypeInt64),
			entity.NewField().WithName("source_file").WithDataType(entity.FieldTypeVarChar).WithMaxLength(256),
		},
		MetricType:        MilvusIndexer.COSINE,
		Embedding:         embedder,
		DocumentConverter: converter,
	}); err != nil {
		slog.Error("milvus indexer init failed", "error", err)
		client.Close()
		return nil
	}

	// 显式设置 SearchParam, 避免 eino-ext 默认的 defaultSearchParam 用向量维度当 radius 导致 COSINE 度量下 "range_filter > radius" 断言失败
	sp, _ := entity.NewIndexAUTOINDEXSearchParam(1)
	if topK <= 0 {
		topK = 10
	}

	// 初始化 Retriever
	retriever, err := MilvusRetriever.NewRetriever(initCtx, &MilvusRetriever.RetrieverConfig{
		Client:            client,
		Collection:        CollectionName,
		OutputFields:      []string{"id", "content", "metadata"},
		DocumentConverter: milvusDocumentConverter,
		VectorConverter:   vectorConverter,
		MetricType:        entity.COSINE,
		TopK:              topK,
		Sp:                sp,
		Embedding:         embedder,
	})
	if err != nil {
		slog.Error("milvus retriever init failed", "error", err)
		client.Close()
		return nil
	}

	return &MilvusStore{
		client:    client,
		embedder:  embedder,
		retriever: retriever,
	}
}

// Store 批量写入题库文档, 返回 int64 主键
func (store *MilvusStore) Store(ctx context.Context, docs []*schema.Document) ([]int64, error) {
	const batchSize = 10
	allIDs := make([]int64, 0)
	if len(docs) == 0 {
		return allIDs, nil
	}
	for i := 0; i < len(docs); i += batchSize {
		end := i + batchSize
		if i+batchSize >= len(docs) {
			end = len(docs)
		}
		batch := docs[i:end]
		texts := make([]string, 0, len(batch))
		for _, doc := range batch {
			texts = append(texts, doc.Content)
		}
		vectors, err := store.embedder.EmbedStrings(ctx, texts)
		if err != nil {
			slog.Error("embed documents failed", "error", err)
			return allIDs, ErrServerInternal
		}
		if len(vectors) != len(batch) {
			slog.Error("embedding result length not match", "want", len(batch), "got", len(vectors))
			return allIDs, ErrServerInternal
		}
		rows, err := converter(ctx, batch, vectors)
		if err != nil {
			return allIDs, err
		}
		inserted, err := store.client.InsertRows(ctx, CollectionName, "", rows)
		if err != nil {
			slog.Error("store failed", "error", err)
			return allIDs, ErrServerInternal
		}
		if err := store.client.Flush(ctx, CollectionName, false); err != nil {
			slog.Error("flush failed", "error", err)
			return allIDs, ErrServerInternal
		}
		for idx := 0; idx < inserted.Len(); idx++ {
			id, err := inserted.GetAsInt64(idx)
			if err != nil {
				slog.Error("get inserted id failed", "error", err)
				return allIDs, ErrServerInternal
			}
			allIDs = append(allIDs, id)
		}
	}
	return allIDs, nil
}

// Retrieve 从 Milvus 检索题目
func (store *MilvusStore) Retrieve(ctx context.Context, query string, opt ...retriever.Option) ([]*schema.Document, error) {
	return store.retriever.Retrieve(ctx, query, opt...)
}

// RetrieveByUser 检索指定用户的题目
func (store *MilvusStore) RetrieveByUser(ctx context.Context, userID int64, query string, opts ...retriever.Option) ([]*schema.Document, error) {
	filter := fmt.Sprintf(`user_id == %d`, userID)
	opts = append(opts, MilvusRetriever.WithFilter(filter))
	return store.retriever.Retrieve(ctx, query, opts...)
}

// DeleteByUserID 删除指定用户的所有题目
func (store *MilvusStore) DeleteByUserID(ctx context.Context, userID int64) error {
	expr := fmt.Sprintf(`user_id == %d`, userID)
	if err := store.client.Delete(ctx, CollectionName, "", expr); err != nil {
		slog.Error("delete by user_id failed", "error", err)
		return ErrServerInternal
	}
	return nil
}

// DeleteBySourceFile 删除指定用户某个来源文件的题目（用于题库更新替换）
func (store *MilvusStore) DeleteBySourceFile(ctx context.Context, userID int64, sourceFile string) error {
	expr := fmt.Sprintf(`user_id == %d && source_file == "%s"`, userID, sourceFile)
	if err := store.client.Delete(ctx, CollectionName, "", expr); err != nil {
		slog.Error("delete by source_file failed", "error", err)
		return ErrServerInternal
	}
	return nil
}

// LoadParsedQuestions 将解析后的结构化题目写入 Milvus（带用户隔离 + 文件来源标记）
func (store *MilvusStore) LoadParsedQuestions(ctx context.Context, userID int64, sourceFile string, questions []model.Question) error {
	docs := make([]*schema.Document, 0, len(questions))
	for _, q := range questions {
		docs = append(docs, &schema.Document{
			// schema.Document.ID 是 string, 业务 ID 在边界处格式化
			ID:      strconv.FormatInt(q.ID, 10),
			Content: q.Content + "\n参考答案：" + q.Reference, // 问题 + 答案拼接利于召回
			MetaData: map[string]any{
				"type":        q.Type,
				"difficulty":  q.Difficulty,
				"skills":      q.Skills,
				"reference":   q.Reference,
				"user_id":     userID,
				"source_file": sourceFile,
			},
		})
	}
	_, err := store.Store(ctx, docs)
	return err
}

// LoadQuestionsFromFile 从 JSON 文件加载面试题到 Milvus（带用户隔离）
func (store *MilvusStore) LoadQuestionsFromFile(ctx context.Context, userID int64, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		slog.Error("read file failed", "error", err)
		return ErrServerInternal
	}

	var questions []model.Question
	if err := json.Unmarshal(data, &questions); err != nil {
		slog.Error("parse file failed", "error", err)
		return ErrServerInternal
	}

	docs := make([]*schema.Document, 0, len(questions))
	for _, q := range questions {
		docs = append(docs, &schema.Document{
			// schema.Document.ID 是 string, 业务 ID 在边界处格式化
			ID:      strconv.FormatInt(q.ID, 10),
			Content: q.Content + "\n参考答案：" + q.Reference,
			MetaData: map[string]any{
				"type":        q.Type,
				"difficulty":  q.Difficulty,
				"skills":      q.Skills,
				"follow_ups":  q.FollowUps,
				"reference":   q.Reference,
				"user_id":     userID,
				"source_file": filePath,
			},
		})
	}

	_, err = store.Store(ctx, docs)
	return err
}
