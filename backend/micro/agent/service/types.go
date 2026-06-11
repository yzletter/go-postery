package service

import "context"

type AgentService interface {
	Chat(ctx context.Context, userID int64, sessionID int64, query string) (string, []string, error)
	StartChunkDocConsumer(ctx context.Context)
	StartUpsertQdrantConsumer(ctx context.Context)
}
