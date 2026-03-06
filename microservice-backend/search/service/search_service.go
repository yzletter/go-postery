package service

import (
	"context"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/segmentio/kafka-go"
	post_grpc "github.com/yzletter/go-postery/api/proto/post/v1"
	"github.com/yzletter/go-postery/client"
	"github.com/yzletter/go-postery/microservice-backend/search/errs"
	model2 "github.com/yzletter/go-postery/microservice-backend/search/model"
	"github.com/yzletter/go-postery/microservice-backend/search/service/index_service"
	ports2 "github.com/yzletter/go-postery/microservice-backend/search/service/ports"
)

type searchService struct {
	indexer       *index_service.Indexer // 单机模式
	kafkaConsumer *kafka.Reader
	tokenizer     ports2.Tokenizer
	idGen         ports2.IDGenerator // 用于生成 ID

	postClient client.PostClient
}

func NewSearchService(kafkaConsumer *kafka.Reader, tokenizer ports2.Tokenizer, idGen ports2.IDGenerator, postClient client.PostClient) SearchService {
	service := &searchService{
		indexer:       new(index_service.Indexer),
		kafkaConsumer: kafkaConsumer,
		tokenizer:     tokenizer,
		idGen:         idGen,
		postClient:    postClient,
	}

	if err := service.indexer.Init(5000000, "data/local_db/search"); err != nil {
		slog.Error("Init Search Index Failed", "error", err)
		return nil
	}
	service.indexer.LoadFromIndexFile() // 从正排中加载数据

	return service
}

func (svc *searchService) Search(ctx context.Context, queries []string) ([]string, error) {
	_ = ctx
	// 构造搜索语句 不同空格之间的应该是和
	var searchQuery = new(model2.TermQuery)
	var titleQuery = new(model2.TermQuery)
	var contentQuery = new(model2.TermQuery)

	for _, query := range queries {
		contentQ := new(model2.TermQuery)
		titleQ := new(model2.TermQuery)
		querySegments := svc.tokenizer.Cut(query)

		for _, segment := range querySegments {
			contentQ = contentQ.And(model2.NewTermQuery("Content", strings.ToLower(segment)))
			titleQ = titleQ.And(model2.NewTermQuery("Title", strings.ToLower(segment)))
		}

		contentQuery = contentQuery.And(contentQ)
		titleQuery = titleQuery.And(titleQ)
	}

	searchQuery = titleQuery.Or(contentQuery)

	// 进行搜索
	documents := svc.indexer.Search(searchQuery, 0, 0, nil)
	docIDs := make([]string, 0, len(documents))
	for _, document := range documents {
		docIDs = append(docIDs, document.DocID)
	}
	return docIDs, nil
}

func (svc *searchService) DeleteDoc(ctx context.Context, docID string) (int, error) {
	_ = ctx
	affectedCount := svc.indexer.DeleteDoc(docID)
	return affectedCount, nil
}

func (svc *searchService) AddDoc(ctx context.Context, doc *model2.Document) (int, error) {
	_ = ctx
	affectedCount, err := svc.indexer.AddDoc(doc)
	if err != nil {
		slog.Error("Server Internal Error", "error", err)
		return affectedCount, errs.ErrInternal
	}
	return affectedCount, nil
}

func (svc *searchService) Count(ctx context.Context) int {
	_ = ctx
	affectedCount := svc.indexer.Count()
	return affectedCount
}

func (svc *searchService) StartConsumer(ctx context.Context) {
	backoff := time.Second
	for {
		select {
		case <-ctx.Done():
			slog.Info("关闭 Index Doc Consumer 成功 ...")
			return
		default:
			// Fetch 消息
			message, err := svc.kafkaConsumer.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil { // 正常退出
					return
				}

				slog.Error("Fetch Message From Kafka Failed", "Kafka", "Search Kafka", "error", err)

				// 简单退避，避免狂刷日志
				time.Sleep(backoff)
				if backoff < 10*time.Second {
					backoff *= 2
				}
				continue
			}

			backoff = time.Second // 重置

			// 解析 JSON
			var payload model2.IndexPayload
			if err = sonic.Unmarshal(message.Value, &payload); err != nil {
				slog.Error("invalid message value, skip", "topic", message.Topic, "partition", message.Partition, "offset", message.Offset, "value", string(message.Value), "errs", err)
				// 脏消息 Commit 掉
				_ = svc.kafkaConsumer.CommitMessages(ctx, message)
				continue
			}

			// 消费消息
			if err = svc.Index(ctx, payload.ID); err != nil {
				slog.Error("Index Failed", "error", err)
				time.Sleep(time.Second) // 最小退避，避免打爆
				continue                // 不 commit -> 重试
			}

			// 消费成功, 把消息 Commit 掉
			if err = svc.kafkaConsumer.CommitMessages(ctx, message); err != nil {
				slog.Error("Commit Kafka Message Failed", "topic", message.Topic, "partition", message.Partition, "offset", message.Offset, "errs", err)
				// Commit 失败通常会导致重复消费，但不会丢消息，可接受
				continue
			}
		}
	}
}

// Index 为新 Post 建立索引
func (svc *searchService) Index(ctx context.Context, postID int64) error {
	// 读文本
	post, err := svc.postClient.GetDetailByID(ctx, &post_grpc.GetDetailByIDRequest{
		PostID:     postID,
		AddViewCnt: false,
	})
	if err != nil {
		slog.Error("Post Service Unavailable", "error", err)
		return errs.ErrUnavailable
	}

	// 对标题分词
	titleSegments := svc.tokenizer.Cut(post.Title)
	titleKeywords := toKeywords("Title", titleSegments)

	// 对文本分词转为关键词
	contentSegments := svc.tokenizer.Cut(post.Content)
	contentKeywords := toKeywords("Content", contentSegments)

	keywords := make([]*model2.Keyword, 0, len(contentSegments)+len(titleSegments))
	keywords = append(keywords, titleKeywords...)
	keywords = append(keywords, contentKeywords...)

	// 建索引
	bs, _ := sonic.Marshal(post) // todo 用 Protobuf
	_, err = svc.indexer.AddDoc(&model2.Document{
		IndexID:     svc.idGen.NextIDUint64(),
		DocID:       strconv.FormatInt(postID, 10),
		BitsFeature: 0, // todo 完善 Post BitsFeature
		Keywords:    keywords,
		Bytes:       bs,
	})
	return nil
}

func toKeywords(field string, segments []string) []*model2.Keyword {
	res := make([]*model2.Keyword, 0, len(segments))

	for _, segment := range segments {
		res = append(res, &model2.Keyword{
			Field: field,
			Word:  strings.ToLower(segment),
		})
	}

	return res
}
