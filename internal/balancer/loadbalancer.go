package balancer

import "sync"

type LoadBalancer struct {
	backends []*Backend
	current  int
	mu       sync.RWMutex
	algo     Algorithm
}

func CreateLB() *LoadBalancer {
	lb := &LoadBalancer{current: 0, mu: sync.RWMutex{}, algo: ROUND_ROBIN}

	lb.backends = make([]*Backend, 4)
	lb.backends[0] = &Backend{ConnType: "tcp", Host: "localhost", Port: "8000", Healthy: true}
	lb.backends[1] = &Backend{ConnType: "tcp", Host: "localhost", Port: "8001", Healthy: false}
	lb.backends[2] = &Backend{ConnType: "tcp", Host: "localhost", Port: "8002", Healthy: true}
	lb.backends[3] = &Backend{ConnType: "tcp", Host: "localhost", Port: "8003", Healthy: true}
	return lb
}
