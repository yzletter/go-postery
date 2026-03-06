package index_service

import (
	model2 "github.com/yzletter/go-postery/microservice-backend/search/model"
)

// IndexerInterface Indexer（单机索引）和 Sentinel（分布式 grpc 的哨兵）都实现了该接口
type IndexerInterface interface {
	AddDoc(doc *model2.Document) (int, error)
	UpdateDoc(doc *model2.Document) (int, error)
	DeleteDoc(docID string) int
	Search(query *model2.TermQuery, onFlag uint64, offFlag uint64, orFlags []uint64) []*model2.Document
	Count() int
	Close() error
}

type LoadBalancer interface {
	Take([]string) string
}
