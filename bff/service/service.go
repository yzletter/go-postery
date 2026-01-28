package service

import (
	"context"
	"log/slog"
	"net/http"

	session_grpc "github.com/yzletter/go-postery/api/proto/session/v1"
	"google.golang.org/grpc"
)

type WebsocketService interface {
	Connect(ctx context.Context, w http.ResponseWriter, r *http.Request, uid int64) error
}

func NewSessionService(endpoint string) session_grpc.SessionServiceClient {
	conn, err := grpc.NewClient(endpoint)
	if err != nil {
		slog.Error("Connect To grpc Failed", "service", "SessionService", "error", err)
		return nil
	}

	client := session_grpc.NewSessionServiceClient(conn)
	return client
}
