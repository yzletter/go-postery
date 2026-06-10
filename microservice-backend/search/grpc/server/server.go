package server

import (
	"context"

	search_grpc "github.com/yzletter/go-postery/api/proto/search/v1"
	"github.com/yzletter/go-postery/microservice-backend/search/model"
	"github.com/yzletter/go-postery/microservice-backend/search/service"
)

type SearchServiceServer struct {
	svc service.SearchService
	search_grpc.UnimplementedSearchServiceServer
}

func NewSearchServiceServer(svc service.SearchService) *SearchServiceServer {
	return &SearchServiceServer{
		svc: svc,
	}
}

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

func (server *SearchServiceServer) DeleteDoc(ctx context.Context, req *search_grpc.DocID) (*search_grpc.AffectedCount, error) {
	count, err := server.svc.DeleteDoc(ctx, req.DocID)
	if err != nil {
		return &search_grpc.AffectedCount{}, err
	}
	return &search_grpc.AffectedCount{Count: int32(count)}, nil
}

func (server *SearchServiceServer) AddDoc(ctx context.Context, req *model.Document) (*search_grpc.AffectedCount, error) {
	count, err := server.svc.AddDoc(ctx, req)
	if err != nil {
		return &search_grpc.AffectedCount{Count: int32(count)}, err
	}
	return &search_grpc.AffectedCount{Count: int32(count)}, nil
}

func (server *SearchServiceServer) Count(ctx context.Context, req *search_grpc.CountRequest) (*search_grpc.AffectedCount, error) {
	_ = req
	count := server.svc.Count(ctx)
	return &search_grpc.AffectedCount{Count: int32(count)}, nil
}

func (server *SearchServiceServer) HealthCheck(ctx context.Context, req *search_grpc.HealthCheckRequest) (*search_grpc.HealthCheckResponse, error) {
	return &search_grpc.HealthCheckResponse{}, nil
}
