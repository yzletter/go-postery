package repository

import (
	"context"

	"github.com/qdrant/go-client/qdrant"
	"github.com/yzletter/go-postery/backend/event"
	"github.com/yzletter/go-postery/backend/micro/agent/model"
)

type AgentRepository interface {
	// Retrieve 检索相关内容
	//
	// Parameter:
	//	- query: 查询内容
	//	- scoreThreshold: 分数阈值
	//	- limit: 查询数量
	//
	// Return:
	//	- []string: 检索结果
	//	- error: 可能返回的错误
	Retrieve(ctx context.Context, query string, scoreThreshold float64, limit int) ([]string, error)

	// CreateChunksWithOutbox 创建文档切块并写入 outbox 事件
	//
	// Parameter:
	//	- chunkModels: 文档切块列表
	//	- event: outbox 事件
	//
	// Return:
	//	- error: 可能返回的错误
	CreateChunksWithOutbox(ctx context.Context, chunkModels []*model.Chunk, event *event.OutboxEvent) error

	// UpsertVectorPoints 写入向量点
	//
	// Parameter:
	//	- points: 向量点列表
	//
	// Return:
	//	- error: 可能返回的错误
	UpsertVectorPoints(ctx context.Context, points []*qdrant.PointStruct) error

	// GetChunksByBatchID 根据批次 ID 获取文档切块
	//
	// Parameter:
	//	- BatchID: 批次 ID
	//
	// Return:
	//	- []*model.Chunk: 文档切块列表
	//	- error: 可能返回的错误
	GetChunksByBatchID(ctx context.Context, BatchID int64) ([]*model.Chunk, error)
}
