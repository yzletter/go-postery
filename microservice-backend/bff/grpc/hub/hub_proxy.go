package hub

import (
	"context"
	"strings"
	"sync"

	etcdv3 "go.etcd.io/etcd/client/v3"
)

type ServiceHubProxy struct {
	*ETCDServiceHub
	watched       sync.Map
	endpointCache sync.Map
}

var (
	ponce sync.Once
	proxy *ServiceHubProxy
)

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
	proxy.watchService(ctx, service)

	if endpoints, exist := proxy.endpointCache.Load(service); exist {
		return endpoints.([]string)
	}

	endpoints := proxy.ETCDServiceHub.GetServiceEndpoints(ctx, service)
	if len(endpoints) > 0 {
		proxy.endpointCache.Store(service, endpoints)
	}
	return endpoints
}

func (proxy *ServiceHubProxy) GetServiceEndpoint(ctx context.Context, service string) string {
	endpoints := proxy.GetServiceEndpoints(ctx, service)
	return proxy.loadBalancer.Take(endpoints)
}

func (proxy *ServiceHubProxy) watchService(ctx context.Context, service string) {
	if _, exists := proxy.watched.LoadOrStore(service, true); exists {
		return
	}

	keyPrefix := proxy.prefix + "/" + service + "/"
	ch := proxy.client.Watch(ctx, keyPrefix, etcdv3.WithPrefix())
	go func() {
		for resp := range ch {
			for _, event := range resp.Events {
				segments := strings.Split(string(event.Kv.Key), "/")
				if len(segments) > 2 {
					service := segments[len(segments)-2]
					endpoints := proxy.ETCDServiceHub.GetServiceEndpoints(ctx, service)
					if len(endpoints) > 0 {
						proxy.endpointCache.Store(service, endpoints)
					} else {
						proxy.endpointCache.Delete(service)
					}
				}
			}
		}
	}()
}
