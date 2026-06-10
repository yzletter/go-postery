package hub

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"

	"github.com/yzletter/go-postery/microservice-backend/user/conf"
	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	etcdv3 "go.etcd.io/etcd/client/v3"
)

// ETCDServiceHub 用 etcd 实现的 ServiceHub
type ETCDServiceHub struct {
	globalMu           sync.RWMutex
	locks              map[string]*sync.RWMutex // 每种服务一把锁
	heartbeatFrequency int64
	prefix             string
	client             *etcdv3.Client       // etcd Client
	loadBalancer       LoadBalancer         // 依赖注入负载均衡器
	pool               map[string]*Endpoint // grpc 连接池
	addrs              map[string][]string
	watched            sync.Map
}

// NewEtcdServiceHub 构造函数
func NewEtcdServiceHub(config conf.ServiceHubConfig, client *etcdv3.Client, loadBalancer LoadBalancer) *ETCDServiceHub {
	return &ETCDServiceHub{
		globalMu:           sync.RWMutex{},
		locks:              make(map[string]*sync.RWMutex),
		heartbeatFrequency: int64(config.HeartbeatFrequency),
		prefix:             config.ServiceRegisterPrefix,
		client:             client,
		loadBalancer:       loadBalancer,
		pool:               make(map[string]*Endpoint),
		addrs:              make(map[string][]string),
	}
}

// LoadEndpoints 初始化建立所有可用连接
func (hub *ETCDServiceHub) LoadEndpoints(ctx context.Context, service string) {
	addrs := hub.getEndpoints(ctx, service)
	for _, addr := range addrs {
		hub.AddEndpoint(ctx, service, addr)
	}
}

// AddEndpoint 向连接池中添加一个 Endpoint
func (hub *ETCDServiceHub) AddEndpoint(ctx context.Context, service string, addr string) {
	// 放入连接池
	lock := hub.getLock(service)
	lock.Lock()
	defer lock.Unlock()

	// 查重
	if _, ok := hub.pool[addr]; ok {
		return
	}

	// 新建
	endpoint := NewEndpoint(addr)

	if hub.addrs[service] == nil {
		hub.addrs[service] = []string{addr}
	} else {
		hub.addrs[service] = append(hub.addrs[service], addr)
	}
	hub.pool[addr] = endpoint
	return
}

// RemoveEndpoint 从连接池中删除一个 Endpoint
func (hub *ETCDServiceHub) RemoveEndpoint(ctx context.Context, service string, addr string) {
	lock := hub.getLock(service)
	lock.Lock()
	defer lock.Unlock()

	endpoint := hub.pool[addr]
	if endpoint != nil {
		if err := endpoint.Close(); err != nil {
			slog.Error("Close Endpoint Failed", "addr", addr, "error", err)
		}
		delete(hub.pool, addr)
	}

	oldAddrs := hub.addrs[service]
	newAddrs := make([]string, 0, len(oldAddrs))
	for _, oldAddr := range oldAddrs {
		if oldAddr != addr {
			newAddrs = append(newAddrs, oldAddr)
		}
	}

	hub.addrs[service] = newAddrs
}

// Take 根据负载均衡算法获取一个可用的 Endpoint todo 有并发问题 ——> Take 完被 Remove
func (hub *ETCDServiceHub) Take(ctx context.Context, service string) *Endpoint {
	lock := hub.getLock(service)
	lock.RLock()
	defer lock.RUnlock()

	// 最多尝试 5 次
	for i := 0; i < len(hub.addrs[service]); i++ {
		endpoint := hub.pool[hub.loadBalancer.Take(service, hub.addrs[service])]
		if endpoint == nil {
			continue
		}
		if endpoint.healthy {
			return endpoint
		}
	}

	return nil
}

func (hub *ETCDServiceHub) WatchEndpointsFromServiceHub(ctx context.Context, service string) {
	// 判断监听过
	if _, exists := hub.watched.LoadOrStore(service, true); exists {
		return
	}

	// 未监听过
	keyPrefix := hub.newPrefix(service)
	ch := hub.client.Watch(ctx, keyPrefix, etcdv3.WithPrefix())

	go func() {
		// 遍历监听管道
		for resp := range ch {
			for _, event := range resp.Events {
				segments := strings.Split(string(event.Kv.Key), "/")
				if len(segments) < 2 {
					// 非法
					continue
				}
				service := segments[len(segments)-2] // 服务名
				addr := segments[len(segments)-1]    // addr
				// 判断是什么操作
				switch event.Type {
				case etcdv3.EventTypePut: // 添加
					hub.AddEndpoint(ctx, service, addr)
				case etcdv3.EventTypeDelete: // 删除
					hub.RemoveEndpoint(ctx, service, addr)
				}
			}
		}
	}()
}

// Register 下游服务向 ServiceHub 注册 / 续约服务
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

	key := hub.newPrefix(service) + endpoint // 构造 Key
	if _, err := hub.client.Put(ctx, key, "", etcdv3.WithLease(resp.ID)); err != nil {
		slog.Error("etcd Put Service Key Failed", "error", err, "Key", key)
		return 0, err
	}

	return int64(resp.ID), nil
}

// Unregister 下游服务向 Service 取消注册服务
//
// service 服务名
//
// endpoint 地址 + 端口
func (hub *ETCDServiceHub) Unregister(ctx context.Context, service string, endpoint string) error {
	// 取消注册
	key := hub.newPrefix(service) + endpoint // 构造 Key
	if _, err := hub.client.Delete(ctx, key); err != nil {
		slog.Error("Unregister Failed", "error", err)
		return err
	}
	return nil
}

// getEndpoints 获得 service 所有节点
func (hub *ETCDServiceHub) getEndpoints(ctx context.Context, service string) []string {
	// 根据前缀查找 Key
	keyPrefix := hub.newPrefix(service)
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

// newPrefix 构造 etcd Key 的前缀
func (hub *ETCDServiceHub) newPrefix(service string) string {
	return hub.prefix + "/" + service + "/"
}

// getLock 获取锁
func (hub *ETCDServiceHub) getLock(service string) *sync.RWMutex {
	// 先用读锁查, 有锁直接返回
	hub.globalMu.RLock()
	lock, ok := hub.locks[service]
	hub.globalMu.RUnlock()
	if ok {
		return lock
	}

	// 没有锁, 再加写锁进行创建
	hub.globalMu.Lock()
	defer hub.globalMu.Unlock()

	// 双重检查，防止其他 goroutine 已经创建了
	if lock, ok := hub.locks[service]; ok {
		return lock
	}

	// 真正创建锁
	lock = &sync.RWMutex{}
	hub.locks[service] = lock
	return lock
}
