package hub

import (
	"log/slog"
	"sync"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

type Endpoint struct {
	mu           sync.RWMutex
	Addr         string
	Conn         *grpc.ClientConn
	failCount    int
	successCount int
	healthy      bool
}

func NewEndpoint(addr string) *Endpoint {
	// 建立连接
	ka := keepalive.ClientParameters{Time: 30 * time.Second, Timeout: 10 * time.Second, PermitWithoutStream: true}
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithKeepaliveParams(ka),
	)
	if err != nil {
		slog.Error("Error creating grpc client connection", "error", err, "addr", addr)
		return nil
	}

	return &Endpoint{
		mu:           sync.RWMutex{},
		Addr:         addr,
		Conn:         conn,
		failCount:    0,
		successCount: 1,
		healthy:      true,
	}
}

func (endpoint *Endpoint) MarkSuccess() {
	endpoint.mu.Lock()
	defer endpoint.mu.Unlock()

	endpoint.failCount = 0
	endpoint.successCount++
	endpoint.healthy = true
}

func (endpoint *Endpoint) MarkFailed() {
	endpoint.mu.Lock()
	defer endpoint.mu.Unlock()

	endpoint.failCount++
	endpoint.successCount = 0
	if endpoint.failCount >= 3 {
		endpoint.healthy = false
	}
}

func (endpoint *Endpoint) IsHealthy() bool {
	endpoint.mu.RLock()
	defer endpoint.mu.RUnlock()
	return endpoint.healthy
}

func (endpoint *Endpoint) Close() error {
	endpoint.mu.Lock()
	defer endpoint.mu.Unlock()
	endpoint.healthy = false
	if endpoint.Conn == nil {
		return nil
	}
	err := endpoint.Conn.Close()
	endpoint.Conn = nil
	return err
}
