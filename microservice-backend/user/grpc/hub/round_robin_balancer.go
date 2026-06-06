package hub

import "sync/atomic"

// RoundRobinLoadBalancer 轮询法负载均衡器
type RoundRobinLoadBalancer struct {
	acc int64
}

func NewRoundRobinLoadBalancer() LoadBalancer {
	return &RoundRobinLoadBalancer{}
}

func (balancer *RoundRobinLoadBalancer) Take(endpoints []string) string {
	length := len(endpoints)
	if length == 0 {
		return ""
	}

	// Take()需要支持并发调用
	n := atomic.AddInt64(&balancer.acc, 1)

	return endpoints[int(n%int64(length))]
}
