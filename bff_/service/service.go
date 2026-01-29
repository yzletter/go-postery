package service

import (
	"context"
	"log/slog"
	"net/http"

	agent_grpc "github.com/yzletter/go-postery/api/proto/agent/v1"
	auth_grpc "github.com/yzletter/go-postery/api/proto/auth/v1"
	code_grpc "github.com/yzletter/go-postery/api/proto/code/v1"
	lottery_grpc "github.com/yzletter/go-postery/api/proto/lottery/v1"
	post_grpc "github.com/yzletter/go-postery/api/proto/post/v1"
	search_grpc "github.com/yzletter/go-postery/api/proto/search/v1"
	session_grpc "github.com/yzletter/go-postery/api/proto/session/v1"
	user_grpc "github.com/yzletter/go-postery/api/proto/user/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type WebsocketService interface {
	Connect(ctx context.Context, w http.ResponseWriter, r *http.Request, uid int64) error
}

func NewSessionService(endpoint string) session_grpc.SessionServiceClient {
	conn, err := grpc.NewClient(
		endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()), // 设置传输安全
	)
	if err != nil {
		slog.Error("Connect To grpc Failed", "service", "SessionService", "error", err)
		return nil
	}

	client := session_grpc.NewSessionServiceClient(conn)
	return client
}
func NewUserService(endpoint string) user_grpc.UserServiceClient {
	conn, err := grpc.NewClient(
		endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()), // 设置传输安全
	)
	if err != nil {
		slog.Error("Connect To grpc Failed", "service", "UserService", "error", err)
		return nil
	}

	client := user_grpc.NewUserServiceClient(conn)
	return client
}

func NewCodeService(endpoint string) code_grpc.CodeServiceClient {
	conn, err := grpc.NewClient(
		endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()), // 设置传输安全
	)
	if err != nil {
		slog.Error("Connect To grpc Failed", "service", "CodeService", "error", err)
		return nil
	}

	client := code_grpc.NewCodeServiceClient(conn)
	return client
}

func NewSearchService(endpoint string) search_grpc.SearchServiceClient {
	conn, err := grpc.NewClient(
		endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()), // 设置传输安全
	)
	if err != nil {
		slog.Error("Connect To grpc Failed", "service", "SearchService", "error", err)
		return nil
	}

	client := search_grpc.NewSearchServiceClient(conn)
	return client
}

func NewPostService(endpoint string) post_grpc.PostServiceClient {
	conn, err := grpc.NewClient(
		endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()), // 设置传输安全
	)
	if err != nil {
		slog.Error("Connect To grpc Failed", "service", "PostService", "error", err)
		return nil
	}

	client := post_grpc.NewPostServiceClient(conn)
	return client
}

func NewLotteryService(endpoint string) lottery_grpc.LotteryServiceClient {
	conn, err := grpc.NewClient(
		endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()), // 设置传输安全
	)
	if err != nil {
		slog.Error("Connect To grpc Failed", "service", "LotteryService", "error", err)
		return nil
	}

	client := lottery_grpc.NewLotteryServiceClient(conn)
	return client
}

func NewAgentService(endpoint string) agent_grpc.AgentServiceClient {
	conn, err := grpc.NewClient(
		endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()), // 设置传输安全
	)
	if err != nil {
		slog.Error("Connect To grpc Failed", "service", "AgentService", "error", err)
		return nil
	}
	w
	client := agent_grpc.NewAgentServiceClient(conn)
	return client
}

func NewAuthService(endpoint string) auth_grpc.AuthServiceClient {
	conn, err := grpc.NewClient(
		endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()), // 设置传输安全
	)
	if err != nil {
		slog.Error("Connect To grpc Failed", "service", "AuthService", "error", err)
		return nil
	}

	client := auth_grpc.NewAuthServiceClient(conn)
	return client
}
