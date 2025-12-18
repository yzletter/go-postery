package ports

type IDGenerator interface {
	NextID() int64
}
