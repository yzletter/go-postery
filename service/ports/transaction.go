package ports

import "context"

type Tx interface {
	DB() any
}

type TransactionManager interface {
	WithTx(ctx context.Context, fn func(ctx context.Context, tx Tx) error) error
}
