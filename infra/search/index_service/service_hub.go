package index_service

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"log/slog"

	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	etcdv3 "go.etcd.io/etcd/client/v3"
)

const (
	ServiceRootPath = "/go-searchery/index"
)

// ServiceHub 服务注册中心
//
// client etcd Client
//
// heartbeatFrequency 心跳上报周期, 单位为秒
//
// loadBalancer 负载均衡算法
type ServiceHub struct {
	client             *etcdv3.Client
	heartbeatFrequency int64
	loadBalancer       LoadBalancer
	watched            sync.Map
}

var (
	serviceHub *ServiceHub // 该全局变量包外不可见，包外想使用时通过 GetServiceHub() 获得
	hubOnce    sync.Once
)

// GetServiceHub 构造函数
//
// etcdServers etcd 监听的地址
//
// heartbeatFrequency etcd 心跳上报周期, 单位为秒
func GetServiceHub(etcdServers []string, heartbeatFrequency int64) *ServiceHub {
	hubOnce.Do(func() {
		client, err := etcdv3.New(etcdv3.Config{Endpoints: etcdServers, DialTimeout: 3 * time.Second})
		if err != nil {
			slog.Error("Init Etcd Failed", "error", err)
			return
		}
		serviceHub = &ServiceHub{
			client:             client,
			heartbeatFrequency: heartbeatFrequency, // 租约的有效期
			loadBalancer:       &RoundRobin{},
		}
	})
	return serviceHub
}

// Register 服务注册
//
// service 服务名
//
// endpoint 微服务 server 的地址
//
// leaseID 租约 ID, 首次注册时置为 0 即可
func (hub *ServiceHub) Register(service string, endpoint string, leaseID etcdv3.LeaseID) (etcdv3.LeaseID, error) {
	ctx := context.Background()
	// 首次注册
	if leaseID <= 0 {
		newLease, err := hub.client.Grant(ctx, hub.heartbeatFrequency)
		if err != nil {
			// 初始化租约失败
			slog.Error("Etcd Grant Lease False", "error", err)
			return 0, err
		}
		// 拼接 Key = ServiceRootPath / service / endpoint
		key := strings.TrimRight(ServiceRootPath, "/") + "/" + service + "/" + endpoint

		// 进行服务注册, 只需要 Key 不需要 Value
		if _, err := hub.client.Put(ctx, key, "", etcdv3.WithLease(newLease.ID)); err != nil {
			slog.Error("Service Register Failed", "error", err)
			return 0, err
		}

		return newLease.ID, nil
	}

	// 续约
	if _, err := hub.client.KeepAliveOnce(ctx, leaseID); err != nil {
		if errors.Is(err, rpctypes.ErrLeaseNotFound) {
			// 租约不存在, 重新注册
			return hub.Register(service, endpoint, 0)
		}
		slog.Error("Keep Lease Failed", "error", err)
		return 0, err
	}

	// 续约成功
	return leaseID, nil
}

// Unregister 服务注销
//
// service 服务名
//
// endpoint 微服务 server 的地址
func (hub *ServiceHub) Unregister(service string, endpoint string) error {
	ctx := context.Background()

	// 拼接 Key = ServiceRootPath + service + endpoint
	key := strings.TrimRight(ServiceRootPath, "/") + "/" + service + "/" + endpoint

	if _, err := hub.client.Delete(ctx, key); err != nil {
		slog.Error("Unregister Failed", "error", err)
		return err
	}

	return nil
}

// GetServiceEndpoints 获取 Service 所有可用 Server
//
// service 服务名
func (hub *ServiceHub) GetServiceEndpoints(service string) []string {
	ctx := context.Background()

	// 拼接前缀
	keyPrefix := strings.TrimRight(ServiceRootPath, "/") + "/" + service + "/"

	// 按前缀查找
	resp, err := hub.client.Get(ctx, keyPrefix, etcdv3.WithPrefix())
	if err != nil {
		slog.Error("Get Service Endpoints Failed", "error", err)
		return nil
	}

	res := make([]string, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		// 只需要 Key, 把 Key 按照 / 切分
		segments := strings.Split(string(kv.Key), "/")

		// 最后一段记录答案
		res = append(res, segments[len(segments)-1])
	}

	return res
}

// GetServiceEndpoint 根据负载均衡获取 Service 一台可用 Server
//
// service 服务名
func (hub *ServiceHub) GetServiceEndpoint(service string) string {
	return hub.loadBalancer.Take(hub.GetServiceEndpoints(service))
}

// Close 关闭连接
func (hub *ServiceHub) Close() {
	_ = hub.client.Close()
}
