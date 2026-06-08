package client

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

var (
	ErrNoAvailableServiceEndpoint = errors.New("no available service endpoint")
	ErrNilConnection              = errors.New("conn is nil")
	ErrNilServiceHub              = errors.New("service hub is nil")
)

type ConnCenter struct {
	serviceHub ServiceHub
}

func NewConnectionCenter(serviceHub ServiceHub) *ConnCenter {
	return &ConnCenter{
		serviceHub: serviceHub,
	}
}

func (center *ConnCenter) NewConnection(ctx context.Context, service string) (*grpc.ClientConn, error) {
	if center == nil || center.serviceHub == nil {
		slog.Error("ServiceHub is nil", "service", service)
		return nil, ErrNilServiceHub
	}

	endpoint := center.serviceHub.GetServiceEndpoint(ctx, service)
	if endpoint == "" {
		slog.Error("No Available Service Endpoint", "service", service)
		return nil, ErrNoAvailableServiceEndpoint
	}

	ka := keepalive.ClientParameters{
		Time:                30 * time.Second,
		Timeout:             10 * time.Second,
		PermitWithoutStream: true,
	}

	return grpc.NewClient(
		endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithKeepaliveParams(ka),
	)
}

func validateConn(conn *grpc.ClientConn) error {
	if conn == nil {
		slog.Error("Conn is nil")
		return ErrNilConnection
	}
	return nil
}
