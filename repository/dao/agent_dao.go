package dao

import (
	"context"
	"errors"
	"log/slog"

	"github.com/cloudwego/eino-ext/components/embedding/ark"
	qdrant_retriever "github.com/cloudwego/eino-ext/components/retriever/qdrant"
	"github.com/go-sql-driver/mysql"
	dao2 "github.com/yzletter/go-postery/lottery/repository/dao"

	"github.com/qdrant/go-client/qdrant"
	"github.com/yzletter/go-postery/model"
	"gorm.io/gorm"
)

type agentDAO struct {
	db          *gorm.DB
	embeddingDB *qdrant.Client
	embedder    *ark.Embedder
}

func NewAgentDAO(db *gorm.DB, embeddingDB *qdrant.Client, embedder *ark.Embedder) AgentDAO {
	return &agentDAO{
		db:          db,
		embeddingDB: embeddingDB,
		embedder:    embedder,
	}
}

// Retrieve 根据 query 召回
func (dao *agentDAO) Retrieve(ctx context.Context, query string, scoreThreshold float64, limit int) ([]string, error) {
	// 构建召回器
	retriever, err := qdrant_retriever.NewRetriever(ctx, &qdrant_retriever.Config{
		Client:         dao.embeddingDB, // 查询数据库需要用到Client
		Embedding:      dao.embedder,    // 把query转成向量需要用到Embedding
		TopK:           limit,           // 召回最相近的几篇文档
		Collection:     "knowledge",     // 表名
		ScoreThreshold: &scoreThreshold,
	})
	if err != nil {
		slog.Error("Qdrant New Retriever Failed", "error", err)
		return nil, dao2.ErrServerInternal
	}

	// 进行召回
	neighbors, err := retriever.Retrieve(ctx, query)
	if err != nil {
		slog.Error("Qdrant Retrieve Failed", "error", err)
		return nil, dao2.ErrServerInternal
	}

	ids := make([]string, 0, len(neighbors))
	for _, neighbor := range neighbors {
		ids = append(ids, neighbor.ID)
	}

	var chunks []*model.Chunk
	result := dao.db.WithContext(ctx).Where("id in ?", ids).Find(&chunks)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return []string{}, nil
		}
		return []string{}, dao2.ErrServerInternal
	}

	res := make([]string, 0, len(neighbors))
	for _, chunk := range chunks {
		res = append(res, chunk.Content)
	}

	return res, nil
}

func (dao *agentDAO) CreateChunksWithOutbox(ctx context.Context, chunkModels []*model.Chunk, event *model.Event) error {
	err := dao.db.WithContext(ctx).Transaction(
		func(tx *gorm.DB) error {
			// 写 Chunk
			if err := tx.CreateInBatches(chunkModels, 200).Error; err != nil {
				return err
			}

			if err := tx.Create(event).Error; err != nil {
				return err
			}
			return nil
		})

	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return dao2.ErrUniqueKey
		}
		slog.Error("Create Chunks Error", "error", err)
		return dao2.ErrServerInternal
	}

	return nil
}

// UpsertVectorPoints 向 Qdrant 中插入向量
func (dao *agentDAO) UpsertVectorPoints(ctx context.Context, points []*qdrant.PointStruct) error {
	// 判断表是否存在
	exist, err := dao.embeddingDB.CollectionExists(ctx, "knowledge")
	if err != nil {
		slog.Error("Qdrant Check Collection Exist Failed", "error", err)
		return dao2.ErrServerInternal
	}
	if !exist {
		// 不存在，建表
		if err := dao.embeddingDB.CreateCollection(ctx, &qdrant.CreateCollection{
			CollectionName: "knowledge",
			VectorsConfig: &qdrant.VectorsConfig{
				Config: &qdrant.VectorsConfig_Params{
					Params: &qdrant.VectorParams{
						Size:     2048,
						Distance: qdrant.Distance_Cosine, // 距离计算方式
					},
				},
			},
		}); err != nil {
			slog.Error("Qdrant Create Collection Failed", "error", err)
			return dao2.ErrServerInternal
		}
	}

	// Upsert
	if _, err := dao.embeddingDB.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: "knowledge",
		Points:         points,
	}); err != nil {
		slog.Error("Qdrant Upsert Vectors Failed", "error", err)
		return dao2.ErrServerInternal
	}

	return nil
}

func (dao *agentDAO) GetChunksByBatchID(ctx context.Context, BatchID int64) ([]*model.Chunk, error) {
	var chunks []*model.Chunk
	result := dao.db.WithContext(ctx).Where("batch_id = ?", BatchID).Find(&chunks)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, dao2.ErrRecordNotFound
		}

		slog.Error(dao2.FindFailed, "batch_id", BatchID, "error", result.Error)
		return nil, dao2.ErrServerInternal
	}

	return chunks, nil
}
