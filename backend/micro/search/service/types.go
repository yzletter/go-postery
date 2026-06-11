package service

import (
	"context"

	"github.com/yzletter/go-postery/backend/micro/search/model"
)

type SearchService interface {
	Search(ctx context.Context, queries []string) ([]string, error)
	DeleteDoc(ctx context.Context, docID string) (int, error)
	AddDoc(ctx context.Context, doc *model.Document) (int, error)
	Count(ctx context.Context) int
	StartConsumer(ctx context.Context)
}
