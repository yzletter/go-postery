package dao

import (
	"context"
	"errors"
	"log/slog"

	"github.com/cloudwego/eino-ext/components/embedding/ark"
	qdrant_retriever "github.com/cloudwego/eino-ext/components/retriever/qdrant"
	"github.com/go-sql-driver/mysql"

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
		return nil, ErrServerInternal
	}

	// 进行召回
	neighbors, err := retriever.Retrieve(ctx, query)
	if err != nil {
		slog.Error("Qdrant Retrieve Failed", "error", err)
		return nil, ErrServerInternal
	}

	res := make([]string, 0, len(neighbors))
	for _, neighbor := range neighbors {
		res = append(res, neighbor.Content)
	}

	return res, nil
}

func (dao *agentDAO) CreateChunks(ctx context.Context, chunkModels []*model.Chunk, event *model.Event) error {
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
			return ErrUniqueKey
		}
		return ErrServerInternal
	}

	return nil
}

func (dao *agentDAO) UpsertVectors(ctx context.Context, chunkModels []*model.Chunk) error {
	points := make([]*qdrant.PointStruct, 0, len(chunkModels))
	for _, chunk := range chunkModels {
		// 用于入 Qdrant
		points = append(points, &qdrant.PointStruct{
			Id:      &qdrant.PointId{PointIdOptions: &qdrant.PointId_Uuid{Uuid: chunk.ID}},
			Vectors: &qdrant.Vectors{VectorsOptions: &qdrant.Vectors_Vector{Vector: &qdrant.Vector{Vector: &qdrant.Vector_Dense{Dense: &qdrant.DenseVector{Data: toFloat32(chunk.Vector)}}}}},
		})
	}

	// 判断表是否存在
	exist, err := dao.embeddingDB.CollectionExists(ctx, "knowledge")
	if err != nil {
		slog.Error("Qdrant Check Collection Exist Failed", "error", err)
		return ErrServerInternal
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
			return ErrServerInternal
		}
	}

	// Upsert
	if _, err := dao.embeddingDB.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: "knowledge",
		Points:         points,
	}); err != nil {
		slog.Error("Qdrant Upsert Vectors Failed", "error", err)
		return ErrServerInternal
	}

	return nil
}

func toFloat32(vector []float64) []float32 {
	rect := make([]float32, len(vector))
	for i, ele := range vector {
		rect[i] = float32(ele)
	}
	return rect
}
