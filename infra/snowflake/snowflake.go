package snowflake

import (
	"log/slog"

	"github.com/bwmarrin/snowflake"
	"github.com/yzletter/go-postery/service/ports"
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
