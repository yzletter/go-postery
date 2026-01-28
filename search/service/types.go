package service

import (
	"context"

	search_grpc "github.com/yzletter/go-postery/api/proto/search/v1"
	"github.com/yzletter/go-postery/search/model"
)

type SearchService interface {
	Search(ctx context.Context, req *search_grpc.SearchRequest) (*search_grpc.SearchResult, error)
	DeleteDoc(ctx context.Context, req *search_grpc.DocID) (*search_grpc.AffectedCount, error)
	AddDoc(ctx context.Context, req *model.Document) (*search_grpc.AffectedCount, error)
	Count(ctx context.Context, req *search_grpc.CountRequest) (*search_grpc.AffectedCount, error)
	StartConsumer(ctx context.Context)
	search_grpc.UnsafeSearchServiceServer
}
