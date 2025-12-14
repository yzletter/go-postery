package snowflake

import (
	"sync"

	"log/slog"

	"github.com/bwmarrin/snowflake"
)

var (
	node *snowflake.Node
	once sync.Once
)

// Init 初始化 Snowflake 节点
func Init(nodeID int) {
	once.Do(func() {
		n, err := snowflake.NewNode(int64(nodeID))
		if err != nil {
			slog.Error("初始化雪花算法失败 ...", "error", err)
		}
		node = n
	})
}

// NextID 生成下一个 ID
func NextID() uint64 {
	if node == nil {
		slog.Error("未初始化雪花算法")
	}
	return uint64(node.Generate())
}
