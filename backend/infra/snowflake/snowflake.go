package snowflake

import (
	"log/slog"
	"math/rand"

	"github.com/bwmarrin/snowflake"
)

type IDGenerator interface {
	NextID() int64
	NextIDUint64() uint64
}

type snowflakeIDGenerator struct {
	node *snowflake.Node
}

func NewSnowflakeIDGenerator(nodeID int) *snowflakeIDGenerator {
	node, err := snowflake.NewNode(int64(nodeID))
	if err != nil {
		slog.Error("Init Snowflake Failed", "error", err)
	}
	return &snowflakeIDGenerator{node: node}
}

func (sf *snowflakeIDGenerator) NextID() int64 {
	if sf.node == nil {
		slog.Error("Snowflake Is Not Initialized")
		return int64(rand.Uint64())
	}
	return int64(sf.node.Generate())
}

func (sf *snowflakeIDGenerator) NextIDUint64() uint64 {
	if sf.node == nil {
		slog.Error("Snowflake Is Not Initialized")
		return rand.Uint64()
	}

	if id := sf.node.Generate(); id > 0 {
		return uint64(id)
	}
	return rand.Uint64()
}
