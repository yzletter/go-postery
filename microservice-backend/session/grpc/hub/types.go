package hub

import (
	"context"
)

type ServiceHub interface {
	Register(ctx context.Context, service string, endpoint string, leaseID int64) (int64, error) // 向 ServiceHub 注册服务
	Unregister(ctx context.Context, service string, endpoint string) error                       // 向 ServiceHub 取消注册服务
	GetServiceEndpoints(ctx context.Context, service string) []string                            // 获得该服务所有可用节点
	GetServiceEndpoint(ctx context.Context, service string) string                               // 负载均衡获得该服务一个可用节点
}

type LoadBalancer interface {
	Take([]string) string // 根据负载均衡算法获得一台可用
}
