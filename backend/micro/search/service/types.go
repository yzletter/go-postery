package service

import (
	"context"

	"github.com/yzletter/go-postery/backend/micro/search/model"
)

type SearchService interface {
	// Search 搜索文档
	//
	// Parameter:
	//	- queries: 查询词列表
	//
	// Return:
	//	- []string: 命中的 DocID 列表
	//	- error: 可能返回的错误
	Search(ctx context.Context, queries []string) ([]string, error)

	// DeleteDoc 删除文档
	//
	// Parameter:
	//	- docID: 文档 ID
	//
	// Return:
	//	- int: 删除数量
	//	- error: 可能返回的错误
	DeleteDoc(ctx context.Context, docID string) (int, error)

	// AddDoc 添加文档
	//
	// Parameter:
	//	- doc: 文档
	//
	// Return:
	//	- int: 添加数量
	//	- error: 可能返回的错误
	AddDoc(ctx context.Context, doc *model.Document) (int, error)

	// Count 获取文档数量
	//
	// Return:
	//	- int: 文档数量
	Count(ctx context.Context) int

	// StartConsumer 启动搜索消息消费者
	//
	// Parameter:
	//	- ctx: 上下文
	StartConsumer(ctx context.Context)
}
