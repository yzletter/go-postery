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
	search_grpc "github.com/yzletter/go-postery/api/proto/search/v1"
	post_conf "github.com/yzletter/go-postery/post/conf"
	"github.com/yzletter/go-postery/search/model"
	"github.com/yzletter/go-postery/search/service/index_service"
	"github.com/yzletter/go-postery/search/service/ports"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type searchService struct {
	indexer       *index_service.Indexer // 单机模式
	kafkaConsumer *kafka.Reader
	tokenizer     ports.Tokenizer
	idGen         ports.IDGenerator // 用于生成 ID
	search_grpc.UnimplementedSearchServiceServer
}

func NewSearchService(kafkaConsumer *kafka.Reader, tokenizer ports.Tokenizer, idGen ports.IDGenerator) SearchService {
	service := &searchService{kafkaConsumer: kafkaConsumer, tokenizer: tokenizer, idGen: idGen}
	service.indexer = new(index_service.Indexer)
	err := service.indexer.Init(5000000, "data/local_db/search")
	if err != nil {
		slog.Error("Init Search Index Failed", "error", err)
		return nil
	}
	service.indexer.LoadFromIndexFile() // 从正排中加载数据
	return service
}

func (svc *searchService) Search(ctx context.Context, req *search_grpc.SearchRequest) (*search_grpc.SearchResult, error) {
	// 构造搜索语句 不同空格之间的应该是和
	var searchQuery = new(model.TermQuery)
	var titleQuery = new(model.TermQuery)
	var contentQuery = new(model.TermQuery)

	for _, query := range req.Queries {
		contentQ := new(model.TermQuery)
		titleQ := new(model.TermQuery)
		querySegments := svc.tokenizer.Cut(query)

		for _, segment := range querySegments {
			contentQ = contentQ.And(model.NewTermQuery("Content", strings.ToLower(segment)))
			titleQ = titleQ.And(model.NewTermQuery("Title", strings.ToLower(segment)))
		}

		contentQuery = contentQuery.And(contentQ)
		titleQuery = titleQuery.And(titleQ)
	}

	searchQuery = titleQuery.Or(contentQuery)

	// 进行搜索
	documents := svc.indexer.Search(searchQuery, 0, 0, nil)
	docIDs := make([]*search_grpc.DocID, 0, len(documents))
	for _, document := range documents {
		docIDs = append(docIDs, &search_grpc.DocID{DocID: document.DocID})
	}
	return &search_grpc.SearchResult{DocumentIDs: docIDs}, nil
}

func (svc *searchService) DeleteDoc(ctx context.Context, req *search_grpc.DocID) (*search_grpc.AffectedCount, error) {
	affectedCount := svc.indexer.DeleteDoc(req.DocID)
	return &search_grpc.AffectedCount{Count: int32(affectedCount)}, nil
}

func (svc *searchService) AddDoc(ctx context.Context, req *model.Document) (*search_grpc.AffectedCount, error) {
	affectedCount, err := svc.indexer.AddDoc(req)
	return &search_grpc.AffectedCount{Count: int32(affectedCount)}, err
}

func (svc *searchService) Count(ctx context.Context, req *search_grpc.CountRequest) (*search_grpc.AffectedCount, error) {
	affectedCount := svc.indexer.Count()
	return &search_grpc.AffectedCount{Count: int32(affectedCount)}, nil
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
			var payload model.IndexPayload
			if err = sonic.Unmarshal(message.Value, &payload); err != nil {
				slog.Error("invalid message value, skip", "topic", message.Topic, "partition", message.Partition, "offset", message.Offset, "value", string(message.Value), "err", err)
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
				slog.Error("Commit Kafka Message Failed", "topic", message.Topic, "partition", message.Partition, "offset", message.Offset, "err", err)
				// Commit 失败通常会导致重复消费，但不会丢消息，可接受
				continue
			}
		}
	}
}

// Index 为新 Post 建立索引
func (svc *searchService) Index(ctx context.Context, postID int64) error {

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

	// 对标题分词
	titleSegments := svc.tokenizer.Cut(post.Title)
	titleKeywords := toKeywords("Title", titleSegments)

	// 对文本分词转为关键词
	contentSegments := svc.tokenizer.Cut(post.Content)
	contentKeywords := toKeywords("Content", contentSegments)

	keywords := make([]*model.Keyword, 0, len(contentSegments)+len(titleSegments))
	keywords = append(keywords, titleKeywords...)
	keywords = append(keywords, contentKeywords...)

	// 建索引
	bs, _ := sonic.Marshal(post) // todo 用 Protobuf
	_, err = svc.indexer.AddDoc(&model.Document{
		IndexID:     svc.idGen.NextIDUint64(),
		DocID:       strconv.FormatInt(postID, 10),
		BitsFeature: 0, // todo 完善 Post BitsFeature
		Keywords:    keywords,
		Bytes:       bs,
	})
	return nil
}

func toKeywords(field string, segments []string) []*model.Keyword {
	res := make([]*model.Keyword, 0, len(segments))

	for _, segment := range segments {
		res = append(res, &model.Keyword{
			Field: field,
			Word:  strings.ToLower(segment),
		})
	}

	return res
}
