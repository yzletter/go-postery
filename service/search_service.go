package service

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/segmentio/kafka-go"
	postdto "github.com/yzletter/go-postery/dto/post"
	infraSearch "github.com/yzletter/go-postery/infra/search/index_service"
	searchModel "github.com/yzletter/go-postery/infra/search/model"
	"github.com/yzletter/go-postery/model"
	"github.com/yzletter/go-postery/repository"
	"github.com/yzletter/go-postery/service/ports"
)

type searchService struct {
	indexer             *infraSearch.Indexer
	searchKafkaConsumer *kafka.Reader
	postRepo            repository.PostRepository
	userRepo            repository.UserRepository
	tokenizer           ports.Tokenizer
	idGen               ports.IDGenerator // 用于生成 ID
}

func NewSearchService(searchKafkaConsumer *kafka.Reader, postRepo repository.PostRepository, userRepo repository.UserRepository, tokenizer ports.Tokenizer, idGen ports.IDGenerator) SearchService {
	service := new(searchService)
	service.tokenizer = tokenizer
	service.searchKafkaConsumer = searchKafkaConsumer
	service.postRepo = postRepo
	service.userRepo = userRepo
	service.idGen = idGen
	service.indexer = new(infraSearch.Indexer)
	err := service.indexer.Init(5000000, "data/local_db/search")
	if err != nil {
		slog.Error("Init Search Index Failed", "error", err)
		return nil
	}
	service.indexer.LoadFromIndexFile() // 从正排中加载数据
	fmt.Println(service.indexer.Count())
	return service
}

func (svc *searchService) StartPostIndexConsumer(ctx context.Context) {
	backoff := time.Second
	for {
		select {
		case <-ctx.Done():
			slog.Info("关闭 Chunk Doc Consumer 成功 ...")
			return
		default:
			// Fetch 消息
			message, err := svc.searchKafkaConsumer.FetchMessage(ctx)
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
			var payload model.ChunkDocumentEventPayload
			err = sonic.Unmarshal(message.Value, &payload)
			if err != nil {
				slog.Error("invalid message value, skip", "topic", message.Topic, "partition", message.Partition, "offset", message.Offset, "value", string(message.Value), "err", err)
				// 脏消息 Commit 掉
				_ = svc.searchKafkaConsumer.CommitMessages(ctx, message)
				continue
			}

			// 消费消息

			err = svc.IndexSearch(ctx, payload.ID)
			if err != nil {
				slog.Error("Index Failed", "error", err)
				time.Sleep(time.Second) // 最小退避，避免打爆
				continue                // 不 commit -> 重试
			}

			// 消费成功, 把消息 Commit 掉
			err = svc.searchKafkaConsumer.CommitMessages(ctx, message)
			if err != nil {
				slog.Error("Commit Kafka Message Failed", "topic", message.Topic, "partition", message.Partition, "offset", message.Offset, "err", err)
				// Commit 失败通常会导致重复消费，但不会丢消息，可接受
				continue
			}
		}
	}
}

func (svc *searchService) IndexSearch(ctx context.Context, id int64) error {
	// 读文本
	post, err := svc.postRepo.GetByID(ctx, id)
	if err != nil {
		slog.Error("获取文本失败", "pid", id, "error", err)
		return err
	}

	// 对标题分词
	titleSegments := svc.tokenizer.Cut(post.Title)
	titleKeywords := toKeywords("Title", titleSegments)

	// 对文本分词转为关键词
	contentSegments := svc.tokenizer.Cut(post.Content)
	contentKeywords := toKeywords("Content", contentSegments)

	keywords := make([]*searchModel.Keyword, 0, len(contentSegments)+len(titleSegments))
	keywords = append(keywords, titleKeywords...)
	keywords = append(keywords, contentKeywords...)

	// 建索引
	bs, _ := sonic.Marshal(post) // todo 用 Protobuf
	_, err = svc.indexer.AddDoc(&searchModel.Document{
		IndexID:     svc.idGen.NextIDUint64(),
		DocID:       strconv.FormatInt(id, 10),
		BitsFeature: 0, // todo 完善 Post BitsFeature
		Keywords:    keywords,
		Bytes:       bs,
	})
	return nil
}

func (svc *searchService) Search(ctx context.Context, querys []string) ([]postdto.DetailDTO, error) {
	// 构造搜索语句 不同空格之间的应该是和
	var searchQuery = new(searchModel.TermQuery)
	var titleQuery = new(searchModel.TermQuery)
	var contentQuery = new(searchModel.TermQuery)

	for _, query := range querys {
		contentQ := new(searchModel.TermQuery)
		titleQ := new(searchModel.TermQuery)
		querySegments := svc.tokenizer.Cut(query)

		for _, segment := range querySegments {
			contentQ = contentQ.And(searchModel.NewTermQuery("Content", strings.ToLower(segment)))
			titleQ = titleQ.And(searchModel.NewTermQuery("Title", strings.ToLower(segment)))
		}

		contentQuery = contentQuery.And(contentQ)
		titleQuery = titleQuery.And(titleQ)
	}

	searchQuery = titleQuery.Or(contentQuery)

	// 进行搜索
	documents := svc.indexer.Search(searchQuery, 0, 0, nil)

	// todo 优化搜索返回，将
	res := make([]postdto.DetailDTO, 0, len(documents))
	for _, document := range documents {
		pid, _ := strconv.ParseInt(document.DocID, 10, 64)

		// 查找帖子
		post, err := svc.postRepo.GetByID(ctx, pid)
		if err != nil {
			continue
		}

		// 查找作者
		user, err := svc.userRepo.GetProfileByID(ctx, post.UserID)
		postDTO := postdto.ToDetailDTO(post, user)
		res = append(res, postDTO)
	}
	return res, nil
}

func toKeywords(field string, segments []string) []*searchModel.Keyword {
	res := make([]*searchModel.Keyword, 0, len(segments))

	for _, segment := range segments {
		res = append(res, &searchModel.Keyword{
			Field: field,
			Word:  strings.ToLower(segment),
		})
	}

	return res
}
