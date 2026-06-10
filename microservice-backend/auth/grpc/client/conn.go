package client

import (
	"context"
	"errors"
	"log/slog"

	"google.golang.org/grpc"
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

	center.serviceHub.LoadEndpoints(ctx, service)
	center.serviceHub.WatchEndpointsFromServiceHub(ctx, service)

	endpoint := center.serviceHub.Take(ctx, service)
	if endpoint == nil || endpoint.Conn == nil {
		slog.Error("No Available Service Endpoint", "service", service)
		return nil, ErrNoAvailableServiceEndpoint
	}

	return endpoint.Conn, nil
}

func validateConn(conn *grpc.ClientConn) error {
	if conn == nil {
		slog.Error("Conn is nil")
		return ErrNilConnection
	}
	return nil
}
