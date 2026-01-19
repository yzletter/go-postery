package dao

import (
	"github.com/qdrant/go-client/qdrant"
	"gorm.io/gorm"
)

type agentDAO struct {
	db          *gorm.DB
	embeddingDB *qdrant.Client
}

func NewAgentDAO(db *gorm.DB, embeddingDB *qdrant.Client) AgentDAO {
	return &agentDAO{
		db:          db,
		embeddingDB: embeddingDB,
	}
}
