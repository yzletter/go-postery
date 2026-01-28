package service

import (
	"context"
	"net/http"
)

type WebsocketService interface {
	Connect(ctx context.Context, w http.ResponseWriter, r *http.Request, uid int64) error
}
