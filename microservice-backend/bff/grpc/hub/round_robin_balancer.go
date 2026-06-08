package hub

import "sync/atomic"

type RoundRobinLoadBalancer struct {
	index uint64
}

func NewRoundRobinLoadBalancer() *RoundRobinLoadBalancer {
	return &RoundRobinLoadBalancer{}
}

func (lb *RoundRobinLoadBalancer) Take(endpoints []string) string {
	if len(endpoints) == 0 {
		return ""
	}

	index := atomic.AddUint64(&lb.index, 1)
	return endpoints[int(index-1)%len(endpoints)]
}
