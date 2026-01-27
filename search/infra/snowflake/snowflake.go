package snowflake

import (
	"log/slog"
	"math/rand"

	"github.com/bwmarrin/snowflake"
	"github.com/yzletter/go-postery/search/service/ports"
)

type snowflakeIDGenerator struct {
	node *snowflake.Node
}

func NewSnowflakeIDGenerator(nodeID int) ports.IDGenerator {
	node, err := snowflake.NewNode(int64(nodeID))
	if err != nil {
		slog.Error("初始化雪花算法失败 ...", "error", err)
	}
	return &snowflakeIDGenerator{node: node}
}

func (sf *snowflakeIDGenerator) NextID() int64 {
	if sf.node == nil {
		slog.Error("未初始化雪花算法")
	}
	return int64(sf.node.Generate())
}

func (sf *snowflakeIDGenerator) NextIDUint64() uint64 {
	if sf.node == nil {
		slog.Error("未初始化雪花算法")
	}

	if id := sf.node.Generate(); id > 0 {
		return uint64(id)
	}
	return rand.Uint64()
}
