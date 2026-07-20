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
	"github.com/yzletter/go-postery/backend/event"
	"github.com/yzletter/go-postery/backend/grpc/errs"
	"github.com/yzletter/go-postery/backend/grpc/manager"
	model2 "github.com/yzletter/go-postery/backend/micro/search/model"
	"github.com/yzletter/go-postery/backend/micro/search/service/index_service"
	"github.com/yzletter/go-postery/backend/ports"
)

type searchService struct {
	indexer       *index_service.Indexer // 单机模式
	kafkaConsumer *kafka.Reader
	tokenizer     ports.Tokenizer
	idGen         ports.IDGenerator // 用于生成 ID

	postClient manager.PostClient
}

func NewSearchService(kafkaConsumer *kafka.Reader, tokenizer ports.Tokenizer, idGen ports.IDGenerator, postClient manager.PostClient) SearchService {
	service := &searchService{
		indexer:       new(index_service.Indexer),
		kafkaConsumer: kafkaConsumer,
		tokenizer:     tokenizer,
		idGen:         idGen,
		postClient:    postClient,
	}

	if err := service.indexer.Init(5000000, "data/local_db/search"); err != nil {
		slog.Error("init search index failed", "error", err)
		return nil
	}
	loadedCount := service.indexer.LoadFromIndexFile() // 从正排中加载历史索引
	slog.Info("search index loaded", "count", loadedCount)

	return service
}

func (svc *searchService) Search(ctx context.Context, queries []string) ([]string, error) {
	_ = ctx
	// 构造搜索语句, 同一字段内多个词取交集, 标题和正文取并集
	var searchQuery = new(model2.TermQuery)
	var titleQuery = new(model2.TermQuery)
	var contentQuery = new(model2.TermQuery)

	for _, query := range queries {
		contentQ := new(model2.TermQuery)
		titleQ := new(model2.TermQuery)
		querySegments := svc.tokenizer.CutSearch(query)

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
	if doc == nil {
		slog.Warn("empty search document, skip")
		return 0, errs.ErrInvalidArgument
	}

	affectedCount, err := svc.indexer.AddDoc(doc)
	if err != nil {
		slog.Error("add search document failed", "doc_id", doc.DocID, "keyword_count", len(doc.Keywords), "error", err)
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
			slog.Info("close search index event consumer success")
			return
		default:
			// Fetch 消息
			message, err := svc.kafkaConsumer.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil { // 正常退出
					return
				}

				slog.Warn("fetch search index event failed", "backoff", backoff, "error", err)

				// 简单退避，避免狂刷日志
				time.Sleep(backoff)
				if backoff < 10*time.Second {
					backoff *= 2
				}
				continue
			}

			backoff = time.Second // 重置

			// 解析 JSON
			var payload event.PostEventPayload
			if err = sonic.Unmarshal(message.Value, &payload); err != nil {
				slog.Warn("invalid search index event, skip", "topic", message.Topic, "partition", message.Partition, "offset", message.Offset, "error", err)
				// 脏消息 Commit 掉
				if commitErr := svc.kafkaConsumer.CommitMessages(ctx, message); commitErr != nil {
					slog.Error("commit invalid search index event failed", "topic", message.Topic, "partition", message.Partition, "offset", message.Offset, "error", commitErr)
				}
				continue
			}

			// 消费消息
			switch payload.EventType {
			case event.PostCreate:
				if err = svc.Index(ctx, payload.ID); err != nil {
					slog.Error("index post failed", "post_id", payload.ID, "error", err)
					time.Sleep(time.Second) // 最小退避，避免打爆
					continue                // 不 commit -> 重试
				}
			case event.PostDelete:
				if _, err = svc.DeleteDoc(ctx, strconv.FormatInt(payload.ID, 10)); err != nil {
					slog.Error("delete post index failed", "post_id", payload.ID, "error", err)
					time.Sleep(time.Second) // 最小退避，避免打爆
					continue                // 不 commit -> 重试
				}
			case event.PostUpdate:
				// 删除旧文档
				_, err = svc.DeleteDoc(ctx, strconv.FormatInt(payload.ID, 10))
				// 添加新文档
				if err = svc.Index(ctx, payload.ID); err != nil {
					slog.Error("index post failed", "post_id", payload.ID, "error", err)
					time.Sleep(time.Second) // 最小退避，避免打爆
					continue                // 不 commit -> 重试
				}
			default:
				// 脏消息 Commit 掉
				if commitErr := svc.kafkaConsumer.CommitMessages(ctx, message); commitErr != nil {
					slog.Error("commit invalid search index event failed", "topic", message.Topic, "partition", message.Partition, "offset", message.Offset, "error", commitErr)
				}
				continue
			}

			// 消费成功, 把消息 Commit 掉
			if err = svc.kafkaConsumer.CommitMessages(ctx, message); err != nil {
				slog.Error("commit search index event failed", "post_id", payload.ID, "topic", message.Topic, "partition", message.Partition, "offset", message.Offset, "error", err)
				// Commit 失败通常会导致重复消费，但不会丢消息，可接受
				continue
			}
		}
	}
}

// Index 为新 Post 建立索引
func (svc *searchService) Index(ctx context.Context, postID int64) error {
	// 查询帖子详情
	post, err := svc.postClient.GetDetailByID(ctx, &post_grpc.GetDetailByIDRequest{
		PostID:     postID,
		AddViewCnt: false,
	})
	if err != nil {
		slog.Error("get post for search index failed", "post_id", postID, "error", err)
		return errs.ErrUnavailable
	}

	// 对标题分词
	titleSegments := svc.tokenizer.CutSearch(post.Title)
	titleKeywords := toKeywords("Title", titleSegments)

	// 对文本分词转为关键词
	contentSegments := svc.tokenizer.CutSearch(post.Content)
	contentKeywords := toKeywords("Content", contentSegments)

	keywords := make([]*model2.Keyword, 0, len(contentSegments)+len(titleSegments))
	keywords = append(keywords, titleKeywords...)
	keywords = append(keywords, contentKeywords...)

	// 序列化原始帖子, 后续查询命中时可直接回源
	bs, err := sonic.Marshal(post) // TODO 后续改为 Protobuf
	if err != nil {
		slog.Error("marshal post for search index failed", "post_id", postID, "error", err)
		return errs.ErrInternal
	}

	// 写入索引
	_, err = svc.indexer.AddDoc(&model2.Document{
		IndexID:     svc.idGen.NextIDUint64(),
		DocID:       strconv.FormatInt(postID, 10),
		BitsFeature: 0, // TODO 后续补充 Post BitsFeature
		Keywords:    keywords,
		Bytes:       bs,
	})
	if err != nil {
		slog.Error("add post to search index failed", "post_id", postID, "keyword_count", len(keywords), "error", err)
		return errs.ErrInternal
	}

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
