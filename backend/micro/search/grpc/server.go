package grpc

import (
	"context"

	search_grpc "github.com/yzletter/go-postery/api/proto/search/v1"
	"github.com/yzletter/go-postery/backend/micro/search/model"
	"github.com/yzletter/go-postery/backend/micro/search/service"
)

type SearchServiceServer struct {
	svc service.SearchService
	search_grpc.UnimplementedSearchServiceServer
}

// NewSearchServiceServer 构造 SearchServiceServer
func NewSearchServiceServer(svc service.SearchService) *SearchServiceServer {
	return &SearchServiceServer{
		svc: svc,
	}
}

// Search 搜索文档
func (server *SearchServiceServer) Search(ctx context.Context, req *search_grpc.SearchRequest) (*search_grpc.SearchResult, error) {
	docIDs, err := server.svc.Search(ctx, req.Queries)
	if err != nil {
		return &search_grpc.SearchResult{}, err
	}

	respIDs := make([]*search_grpc.DocID, 0, len(docIDs))
	for _, docID := range docIDs {
		respIDs = append(respIDs, &search_grpc.DocID{DocID: docID})
	}

	return &search_grpc.SearchResult{DocumentIDs: respIDs}, nil
}

// DeleteDoc 删除文档索引
func (server *SearchServiceServer) DeleteDoc(ctx context.Context, req *search_grpc.DocID) (*search_grpc.AffectedCount, error) {
	count, err := server.svc.DeleteDoc(ctx, req.DocID)
	if err != nil {
		return &search_grpc.AffectedCount{}, err
	}
	return &search_grpc.AffectedCount{Count: int32(count)}, nil
}

// AddDoc 添加文档索引
func (server *SearchServiceServer) AddDoc(ctx context.Context, req *search_grpc.Document) (*search_grpc.AffectedCount, error) {
	count, err := server.svc.AddDoc(ctx, toModelDocument(req))
	if err != nil {
		return &search_grpc.AffectedCount{Count: int32(count)}, err
	}
	return &search_grpc.AffectedCount{Count: int32(count)}, nil
}

// Count 获取索引文档数量
func (server *SearchServiceServer) Count(ctx context.Context, req *search_grpc.CountRequest) (*search_grpc.AffectedCount, error) {
	_ = req
	count := server.svc.Count(ctx)
	return &search_grpc.AffectedCount{Count: int32(count)}, nil
}

// HealthCheck 健康检查
func (server *SearchServiceServer) HealthCheck(ctx context.Context, req *search_grpc.HealthCheckRequest) (*search_grpc.HealthCheckResponse, error) {
	return &search_grpc.HealthCheckResponse{}, nil
}

// toModelDocument 转换 gRPC Document
func toModelDocument(doc *search_grpc.Document) *model.Document {
	if doc == nil {
		return nil
	}

	keywords := make([]*model.Keyword, 0, len(doc.Keywords))
	for _, keyword := range doc.Keywords {
		if keyword == nil {
			continue
		}
		keywords = append(keywords, &model.Keyword{
			Field: keyword.Field,
			Word:  keyword.Word,
		})
	}

	return &model.Document{
		IndexID:     doc.IndexID,
		DocID:       doc.DocID,
		BitsFeature: doc.BitsFeature,
		Keywords:    keywords,
		Bytes:       doc.Bytes,
	}
}
