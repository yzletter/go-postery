package hub

import (
	"context"
	"strings"
	"sync"

	etcdv3 "go.etcd.io/etcd/client/v3"
)

// ServiceHubProxy 额外提供缓存服务的 ServiceHub 代理
type ServiceHubProxy struct {
	*ETCDServiceHub
	watched       sync.Map
	endpointCache sync.Map // 缓存
}

var (
	ponce sync.Once
	proxy *ServiceHubProxy
)

// GetServiceHubProxy 单例模式构造函数
func GetServiceHubProxy(ServiceHub *ETCDServiceHub) *ServiceHubProxy {
	ponce.Do(func() {
		proxy = &ServiceHubProxy{
			ETCDServiceHub: ServiceHub,
			endpointCache:  sync.Map{},
		}
	})
	return proxy
}

func (proxy *ServiceHubProxy) GetServiceEndpoints(ctx context.Context, service string) []string {
	// 先监听
	proxy.watchService(ctx, service)

	// 先查缓存
	if endpoints, exist := proxy.endpointCache.Load(service); exist {
		return endpoints.([]string)
	}

	// 缓存没有, 查 etcd
	endpoints := proxy.ETCDServiceHub.GetServiceEndpoints(ctx, service)
	if len(endpoints) > 0 {
		proxy.endpointCache.Store(service, endpoints) // 查询 etcd 的结果放入缓存
	}
	return endpoints

}

func (proxy *ServiceHubProxy) GetServiceEndpoint(ctx context.Context, service string) string {
	endpoints := proxy.GetServiceEndpoints(ctx, service)
	return proxy.loadBalancer.Take(endpoints)
}

func (proxy *ServiceHubProxy) watchService(ctx context.Context, service string) {
	// 判断监听过
	if _, exists := proxy.watched.LoadOrStore(service, true); exists {
		return
	}

	// 未监听过
	keyPrefix := proxy.prefix + "/" + service + "/"
	ch := proxy.client.Watch(ctx, keyPrefix, etcdv3.WithPrefix())
	go func() {
		// 遍历监听管道
		for resp := range ch {
			for _, event := range resp.Events {
				segments := strings.Split(string(event.Kv.Key), "/")
				if len(segments) > 2 {
					service := segments[len(segments)-2] // 服务名
					// 同步一次 etcd
					endpoints := proxy.ETCDServiceHub.GetServiceEndpoints(ctx, service)
					if len(endpoints) > 0 {
						proxy.endpointCache.Store(service, endpoints) // 查询 etcd 的结果放入缓存
					} else {
						proxy.endpointCache.Delete(service) // 该 service 下已经没有 endpoint, 从缓存中删除
					}
				}
			}
		}
	}()

}
