package hub

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/yzletter/go-postery/microservice-backend/search/config"
	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	etcdv3 "go.etcd.io/etcd/client/v3"
)

// ETCDServiceHub 用 etcd 实现的 ServiceHub
type ETCDServiceHub struct {
	heartbeatFrequency int64
	prefix             string
	client             *etcdv3.Client // etcd Client
	loadBalancer       LoadBalancer   // 依赖注入负载均衡器
}

// NewEtcdServiceHub 构造函数
func NewEtcdServiceHub(config config.ServiceHubConfig, client *etcdv3.Client, loadBalancer LoadBalancer) *ETCDServiceHub {
	return &ETCDServiceHub{
		heartbeatFrequency: int64(config.HeartbeatFrequency),
		prefix:             config.ServiceRegisterPrefix,
		client:             client,
		loadBalancer:       loadBalancer,
	}
}

// Register 向 Service 注册 / 续约服务
//
// service 服务名
//
// endpoint 地址 + 端口
//
// leaseID 租约, 首次注册时置为 0
func (hub *ETCDServiceHub) Register(ctx context.Context, service string, endpoint string, leaseID int64) (int64, error) {
	// 续约
	if leaseID > 0 {
		if _, err := hub.client.KeepAliveOnce(ctx, etcdv3.LeaseID(leaseID)); err != nil {
			if errors.Is(err, rpctypes.ErrLeaseNotFound) {
				// 租约不存在, 重新注册
				return hub.Register(ctx, service, endpoint, 0)
			}
			// 续约失败
			slog.Error("Register Service Failed", "error", err, "leaseID", leaseID)
			return 0, err
		}
		// 续约成功
		return leaseID, nil
	}

	// 首次注册
	resp, err := hub.client.Grant(ctx, hub.heartbeatFrequency) // 生成租约
	if err != nil {
		slog.Error("Register Service Failed", "error", err)
		return 0, err
	}

	key := hub.prefix + "/" + service + "/" + endpoint // 构造 Key
	if _, err := hub.client.Put(ctx, key, "", etcdv3.WithLease(resp.ID)); err != nil {
		slog.Error("etcd Put Service Key Failed", "error", err, "Key", key)
		return 0, err
	}

	return int64(resp.ID), nil
}

// Unregister 向 Service 取消注册服务
//
// service 服务名
//
// endpoint 地址 + 端口
func (hub *ETCDServiceHub) Unregister(ctx context.Context, service string, endpoint string) error {
	// 取消注册
	key := hub.prefix + "/" + service + "/" + endpoint // 构造 Key
	if _, err := hub.client.Delete(ctx, key); err != nil {
		slog.Error("Unregister Failed", "error", err)
		return err
	}
	return nil
}

// GetServiceEndpoints 获取当前服务所有可用节点
//
// service 服务名
func (hub *ETCDServiceHub) GetServiceEndpoints(ctx context.Context, service string) []string {
	// 根据前缀查找 Key
	keyPrefix := hub.prefix + "/" + service + "/"
	resp, err := hub.client.Get(ctx, keyPrefix, etcdv3.WithPrefix())
	if err != nil {
		slog.Error("Get Service Endpoints Failed", "error", err)
		return nil
	}

	// 遍历 KV
	endpoints := make([]string, len(resp.Kvs))
	for i, kv := range resp.Kvs {
		// 按 "/" 切分 Key
		segments := strings.Split(string(kv.Key), "/")
		endpoints[i] = segments[len(segments)-1] // 只要最后一段
	}

	return endpoints
}

// GetServiceEndpoint 根据负载均衡算法获取当前服务一台可用节点
//
// service 服务名
func (hub *ETCDServiceHub) GetServiceEndpoint(ctx context.Context, service string) string {
	return hub.loadBalancer.Take(hub.GetServiceEndpoints(ctx, service))
}
