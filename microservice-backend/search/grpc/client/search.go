package client

import (
	"context"
	"time"

	search_grpc "github.com/yzletter/go-postery/api/proto/search/v1"
	search_model "github.com/yzletter/go-postery/microservice-backend/search/model"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

type searchClient struct {
	conn   *grpc.ClientConn
	client search_grpc.SearchServiceClient
}

func NewSearchClient() (SearchClient, error) {
	// 建议：启用 ka，避免中间网络设备把长连接静默掐掉
	ka := keepalive.ClientParameters{
		Time:                30 * time.Second,
		Timeout:             10 * time.Second,
		PermitWithoutStream: true,
	}

	conn, err := grpc.NewClient(
		SearchClientAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()), // 生产用 TLS
		CircuitBreakerDialOption(),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()), // Jaeger
		grpc.WithKeepaliveParams(ka),
	)
	if err != nil {
		return nil, err
	}

	return &searchClient{
		conn:   conn,
		client: search_grpc.NewSearchServiceClient(conn),
	}, nil
}

func (client *searchClient) Close() {
	_ = client.conn.Close()
}

func (client *searchClient) Search(ctx context.Context, req *search_grpc.SearchRequest) (*search_grpc.SearchResult, error) {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(ctx, 8000*time.Millisecond)
	defer cancel()

	return client.client.Search(ctx, req)
}

func (client *searchClient) DeleteDoc(ctx context.Context, req *search_grpc.DocID) (*search_grpc.AffectedCount, error) {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(ctx, 8000*time.Millisecond)
	defer cancel()

	return client.client.DeleteDoc(ctx, req)
}

func (client *searchClient) AddDoc(ctx context.Context, req *search_model.Document) (*search_grpc.AffectedCount, error) {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(ctx, 8000*time.Millisecond)
	defer cancel()

	return client.client.AddDoc(ctx, req)
}

func (client *searchClient) Count(ctx context.Context, req *search_grpc.CountRequest) (*search_grpc.AffectedCount, error) {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(ctx, 8000*time.Millisecond)
	defer cancel()

	return client.client.Count(ctx, req)
}
