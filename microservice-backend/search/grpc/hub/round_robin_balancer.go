package hub

import (
	"sync"
	"sync/atomic"
)

// RoundRobinLoadBalancer 轮询法负载均衡器
type RoundRobinLoadBalancer struct {
	globalMu sync.RWMutex
	acc      map[string]*uint64
}

func NewRoundRobinLoadBalancer() LoadBalancer {
	return &RoundRobinLoadBalancer{
		acc: make(map[string]*uint64),
	}
}

// Take 根据负载均衡算法选取一个节点
func (balancer *RoundRobinLoadBalancer) Take(service string, endpoints []string) string {
	length := len(endpoints)
	if length == 0 {
		return ""
	}

	counter := balancer.getCounter(service)

	// Take() 需要支持并发调用
	// AddUint64 返回自增后的值，所以减 1，让第一次命中 endpoints[0]
	n := atomic.AddUint64(counter, 1) - 1

	return endpoints[int(n%uint64(length))]
}

// 获取 service 对应计数器
func (balancer *RoundRobinLoadBalancer) getCounter(service string) *uint64 {
	// 先用读锁查
	balancer.globalMu.RLock()
	counter, ok := balancer.acc[service]
	balancer.globalMu.RUnlock()

	if ok {
		return counter
	}

	// 没有则加写锁创建
	balancer.globalMu.Lock()
	defer balancer.globalMu.Unlock()

	// 双重检查，防止其他 goroutine 已经创建
	if counter, ok := balancer.acc[service]; ok {
		return counter
	}

	counter = new(uint64)
	balancer.acc[service] = counter

	return counter
}
