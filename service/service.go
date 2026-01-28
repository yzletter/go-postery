package service

import (
	"context"
	"net/http"

	agentdto "github.com/yzletter/go-postery/dto/agent"
	messagedto "github.com/yzletter/go-postery/dto/message"
	sessiondto "github.com/yzletter/go-postery/dto/session"
	userdto "github.com/yzletter/go-postery/dto/user"
	giftdto "github.com/yzletter/go-postery/lottery/dto/gift"
	orderdto "github.com/yzletter/go-postery/lottery/dto/order"
	"github.com/yzletter/go-postery/service/ports"
)

// 定义 Service 层所有接口

type SessionService interface {
	ListByUid(ctx context.Context, uid int64) ([]sessiondto.DTO, error)
	GetSession(ctx context.Context, uid, targetID int64) (sessiondto.DTO, error)
	Register(ctx context.Context, uid int64) error
	GetHistoryMessagesByPage(ctx context.Context, uid int64, targetID int64, pageNo, pageSize int) (int, []messagedto.DTO, error)
	Delete(ctx context.Context, uid, sid int64) error
	StartSessionRegisterConsumer(ctx context.Context)
}

type WebsocketService interface {
	Connect(ctx context.Context, w http.ResponseWriter, r *http.Request, uid int64) error
}
