package main

import "sync/atomic"

type Strategy interface {
	Select(lb *TCPBalancer) *Backend
}

type RoundRobinStrategy struct{}

func NewRoundRobinStrategy() *RoundRobinStrategy { return &RoundRobinStrategy{} }

func (r *RoundRobinStrategy) Select(lb *TCPBalancer) *Backend {
	alive := lb.GetAliveBackends()
	n := len(alive)
	if n == 0 {
		return nil
	}
	i := int(atomic.AddUint64(&lb.rrCounter, 1) - 1)
	return alive[i%n]
}

type LeastConnStrategy struct{}

func NewLeastConnStrategy() *LeastConnStrategy { return &LeastConnStrategy{} }

func (l *LeastConnStrategy) Select(lb *TCPBalancer) *Backend {
	alive := lb.GetAliveBackends()
	if len(alive) == 0 {
		return nil
	}
	var sel *Backend
	min := int64(1<<62 - 1)
	for _, b := range alive {
		if c := b.ActiveConns(); c < min {
			min = c
			sel = b
		}
	}
	return sel
}
