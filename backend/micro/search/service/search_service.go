package service

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"

	"github.com/bytedance/sonic"
	"github.com/segmentio/kafka-go"
	"github.com/yzletter/go-postery/backend/api/proto/post/v1"
	"github.com/yzletter/go-postery/backend/grpc/errs"
	"github.com/yzletter/go-postery/backend/grpc/manager"
	"github.com/yzletter/go-postery/backend/micro/outbox/model"
	outboxService "github.com/yzletter/go-postery/backend/micro/outbox/service"
	model2 "github.com/yzletter/go-postery/backend/micro/search/model"
	"github.com/yzletter/go-postery/backend/micro/search/service/index_service"
	"github.com/yzletter/go-postery/backend/ports"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type searchService struct {
	indexMu       sync.Mutex
	indexer       *index_service.Indexer // 单机模式
	kafkaConsumer *outboxService.Consumer
	tokenizer     ports.Tokenizer
	idGen         ports.IDGenerator // 用于生成 ID

	postClient manager.PostClient
}

func NewSearchService(kafkaConsumer *outboxService.Consumer, tokenizer ports.Tokenizer, idGen ports.IDGenerator, postClient manager.PostClient) SearchService {
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
	return svc.indexer.DeleteDocChecked(docID)
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
	svc.kafkaConsumer.Run(ctx, svc.consumeIndexEvent, nil)
}

// consumeIndexEvent 根据帖子事件更新或删除搜索索引
func (svc *searchService) consumeIndexEvent(ctx context.Context, messages ...kafka.Message) error {
	var payload model.PostEventPayload
	if err := outboxService.DecodeMessage(messages[0], &payload); err != nil {
		return err
	}
	if payload.ID <= 0 {
		return fmt.Errorf("%w: invalid post id", outboxService.ErrInvalidMessage)
	}

	// 同一个帖子可能同时收到正常消息和人工回放。
	svc.indexMu.Lock()
	defer svc.indexMu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	switch payload.EventType {
	case model.PostCreate, model.PostUpdate:
		return svc.Index(ctx, payload.ID)
	case model.PostDelete:
		_, err := svc.DeleteDoc(ctx, strconv.FormatInt(payload.ID, 10))
		return err
	default:
		return fmt.Errorf("%w: unknown search event type %q", outboxService.ErrInvalidMessage, payload.EventType)
	}
}

// Index 为新 Post 建立索引
func (svc *searchService) Index(ctx context.Context, postID int64) error {
	// 查询帖子详情
	post, err := svc.postClient.GetDetailByID(ctx, &post_grpc.GetDetailByIDRequest{
		PostID:     postID,
		AddViewCnt: false,
	})
	
	if status.Code(err) == codes.NotFound {
		// 回放旧事件时，按帖子当前状态移除已删除帖子的索引。
		_, err := svc.DeleteDoc(ctx, strconv.FormatInt(postID, 10))
		return err
	}
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
	_, err = svc.indexer.UpdateDoc(&model2.Document{
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
