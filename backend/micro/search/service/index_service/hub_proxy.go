package index_service

import (
	"context"
	"strings"
	"sync"
	"time"

	etcdv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/time/rate"
)

// ServiceHubInterface 服务注册中心接口, 代理通过组合实现该接口
type ServiceHubInterface interface {
	// Register 注册服务
	//
	// Parameter:
	//	- service: 服务名
	//	- endpoint: 服务节点
	//	- leaseID: 租约 ID
	//
	// Return:
	//	- etcdv3.LeaseID: 租约 ID
	//	- error: 可能返回的错误
	Register(service string, endpoint string, leaseID etcdv3.LeaseID) (etcdv3.LeaseID, error)

	// Unregister 注销服务
	//
	// Parameter:
	//	- service: 服务名
	//	- endpoint: 服务节点
	//
	// Return:
	//	- error: 可能返回的错误
	Unregister(service string, endpoint string) error

	// GetServiceEndpoints 服务发现
	//
	// Parameter:
	//	- service: 服务名
	//
	// Return:
	//	- []string: 服务节点列表
	GetServiceEndpoints(service string) []string

	// GetServiceEndpoint 选择服务的一个 endpoint
	//
	// Parameter:
	//	- service: 服务名
	//
	// Return:
	//	- string: 服务节点
	GetServiceEndpoint(service string) string

	// Close 关闭 etcd Client 连接
	Close()
}

// ServiceHubProxy 代理模式, 提供缓存和限流能力
//
// endpointCache 维护每个 service 下的所有 endpoint
//
// limiter 限流服务
type ServiceHubProxy struct {
	*ServiceHub
	endpointCache sync.Map
	limiter       *rate.Limiter
}

var (
	proxy     *ServiceHubProxy
	proxyOnce sync.Once
)

// GetServiceHubProxy 构造函数
//
// etcdServers etcd 监听地址
//
// heartbeatFrequency etcd 心跳上报周期, 单位为秒
//
// qps 每秒产生 qps 个令牌, 令牌桶容量也为 qps
func GetServiceHubProxy(etcdServers []string, heartbeatFrequency int64, qps int) *ServiceHubProxy {
	proxyOnce.Do(func() {
		serviceHub := GetServiceHub(etcdServers, heartbeatFrequency)
		if serviceHub == nil {
			return
		}

		proxy = new(ServiceHubProxy)
		proxy.ServiceHub = serviceHub
		proxy.endpointCache = sync.Map{}
		proxy.limiter = rate.NewLimiter(rate.Every(time.Duration(1e9/qps)*time.Nanosecond), qps)
	})

	return proxy
}

// GetServiceEndpoints 服务发现
//
// 第一次查询 etcd 后写入缓存, 后续通过 Watch 更新缓存, 降低 etcd 访问压力
//
// service 服务名
func (proxy *ServiceHubProxy) GetServiceEndpoints(service string) []string {
	// 判断是否限流
	if !proxy.limiter.Allow() {
		return nil
	}

	// 安装 Watch
	proxy.watchEndpointsOfService(service)

	// 先查缓存
	if endpoints, exist := proxy.endpointCache.Load(service); exist {
		return endpoints.([]string)
	}

	endpoints := proxy.ServiceHub.GetServiceEndpoints(service)
	if len(endpoints) > 0 {
		proxy.endpointCache.Store(service, endpoints) // 查询 etcd 的结果放入缓存
	}
	return endpoints
}

func (proxy *ServiceHubProxy) watchEndpointsOfService(service string) {
	ctx := context.Background()
	if _, exists := proxy.watched.LoadOrStore(service, true); exists { // watched 从父类继承
		return // 监听过了, 不用重复监听
	}

	// 拼接前缀
	keyPrefix := strings.TrimRight(ServiceRootPath, "/") + "/" + service + "/"

	// 根据前缀监听, 每次修改都会放入 ch
	ch := proxy.client.Watch(ctx, keyPrefix, etcdv3.WithPrefix())

	go func() {
		// 遍历 ch
		for response := range ch {
			for _, event := range response.Events {
				// 只需要 Key, 把 Key 按 / 切分
				segments := strings.Split(string(event.Kv.Key), "/")
				if len(segments) > 2 {
					service := segments[len(segments)-2] // 倒数第二个是服务名
					// 同步一次 etcd
					endpoints := proxy.ServiceHub.GetServiceEndpoints(service)
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
