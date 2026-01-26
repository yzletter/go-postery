package service

import (
	"context"
	"log/slog"

	infraSearch "github.com/yzletter/go-postery/infra/search/index_service"
)

type searchService struct {
	indexer *infraSearch.Indexer
}

func NewSearchService() SearchService {
	service := new(searchService)
	service.indexer = new(infraSearch.Indexer)
	err := service.indexer.Init(5000000, "data/local_db/search")
	if err != nil {
		slog.Error("Init Search Index Failed", "error", err)
		return nil
	}
	service.indexer.LoadFromIndexFile() // 从正排中加载数据
	return service
}

func (svc *searchService) StartPostIndexConsumer(ctx context.Context) {
	//TODO implement me
	panic("implement me")
}
