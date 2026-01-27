package index_service

import "github.com/yzletter/go-postery/search/model"

// IndexerInterface Indexer（单机索引）和 Sentinel（分布式 grpc 的哨兵）都实现了该接口
type IndexerInterface interface {
	AddDoc(doc *model.Document) (int, error)
	UpdateDoc(doc *model.Document) (int, error)
	DeleteDoc(docID string) int
	Search(query *model.TermQuery, onFlag uint64, offFlag uint64, orFlags []uint64) []*model.Document
	Count() int
	Close() error
}

type LoadBalancer interface {
	Take([]string) string
}
