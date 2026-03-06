package ports

type IDGenerator interface {
	NextID() int64
	NextIDUint64() uint64
}
