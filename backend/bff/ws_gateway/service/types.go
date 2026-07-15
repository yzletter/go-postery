package service

import (
	"context"

	"github.com/yzletter/go-postery/backend/bff/ws_gateway/domain"
)

// WSMessageHandler 用来处理客户端发来的业务消息
type WSMessageHandler interface {
	// NewSessionConnection 处理新 Session 连接
	NewSessionConnection(ctx context.Context, userID int64) error
	// HandleWSMessage 处理客户端发来的业务消息。
	HandleWSMessage(ctx context.Context, userID int64, biz domain.ConnType, msg WSMessage) error
}
