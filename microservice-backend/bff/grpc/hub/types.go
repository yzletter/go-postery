package hub

import (
	"context"
)

type ServiceHub interface {
	Register(ctx context.Context, service string, endpoint string, leaseID int64) (int64, error)
	Unregister(ctx context.Context, service string, endpoint string) error
	GetServiceEndpoints(ctx context.Context, service string) []string
	GetServiceEndpoint(ctx context.Context, service string) string
}

type LoadBalancer interface {
	Take([]string) string
}
