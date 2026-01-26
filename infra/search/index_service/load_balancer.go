package index_service

import "sync/atomic"

// RoundRobin 负载均衡算法--轮询法
type RoundRobin struct {
	acc int64
}

func (b *RoundRobin) Take(endpoints []string) string {
	if len(endpoints) == 0 {
		return ""
	}
	n := atomic.AddInt64(&b.acc, 1) // Take()需要支持并发调用
	index := int(n % int64(len(endpoints)))
	return endpoints[index]
}
