package hub

import (
	"context"
)

type ServiceHub interface {
	LoadEndpoints(ctx context.Context, service string)                                           // 从服务注册中心初始化所有可用连接
	AddEndpoint(ctx context.Context, service string, addr string)                                // 建立新连接
	RemoveEndpoint(ctx context.Context, service string, addr string)                             // 删除节点连接
	Take(ctx context.Context, service string) *Endpoint                                          // 根据负载均衡选择一个连接
	WatchEndpointsFromServiceHub(ctx context.Context, service string)                            // Watch 一个服务
	Register(ctx context.Context, service string, endpoint string, leaseID int64) (int64, error) // 下游服务向 ServiceHub 注册 / 续约服务
	Unregister(ctx context.Context, service string, endpoint string) error                       // 下游服务向 ServiceHub 取消注册服务
}

type LoadBalancer interface {
	Take(service string, endpoints []string) string // 根据负载均衡算法获得一台可用
}
