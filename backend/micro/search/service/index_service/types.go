package index_service

import (
	model2 "github.com/yzletter/go-postery/backend/micro/search/model"
)

// IndexerInterface Indexer（单机索引）和 Sentinel（分布式 grpc 的哨兵）都实现了该接口
type IndexerInterface interface {
	// AddDoc 添加文档
	//
	// Parameter:
	//	- doc: 文档
	//
	// Return:
	//	- int: 添加数量
	//	- error: 可能返回的错误
	AddDoc(doc *model2.Document) (int, error)

	// UpdateDoc 更新文档
	//
	// Parameter:
	//	- doc: 文档
	//
	// Return:
	//	- int: 更新数量
	//	- error: 可能返回的错误
	UpdateDoc(doc *model2.Document) (int, error)

	// DeleteDoc 删除文档
	//
	// Parameter:
	//	- docID: 文档 ID
	//
	// Return:
	//	- int: 删除数量
	DeleteDoc(docID string) int

	// Search 搜索文档
	//
	// Parameter:
	//	- query: 查询条件
	//	- onFlag: 必须开启的标志
	//	- offFlag: 必须关闭的标志
	//	- orFlags: 任一匹配标志列表
	//
	// Return:
	//	- []*model2.Document: 文档列表
	Search(query *model2.TermQuery, onFlag uint64, offFlag uint64, orFlags []uint64) []*model2.Document

	// Count 获取文档数量
	//
	// Return:
	//	- int: 文档数量
	Count() int

	// Close 关闭索引
	//
	// Return:
	//	- error: 可能返回的错误
	Close() error
}

type LoadBalancer interface {
	// Take 选择一个服务节点
	//
	// Parameter:
	//	- []string: 服务节点列表
	//
	// Return:
	//	- string: 服务节点
	Take([]string) string
}
