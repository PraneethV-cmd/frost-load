package pool

import (
	"errors"
	"sync/atomic"
)

/*
Notes: atomics cannot be copied, so dont use func (b backend) as this makes or does global copying


*/

type Backend struct {
	addr        string
	weight      int
	alive       atomic.Bool
	activeConns atomic.Int64
}

func NewBackend(addr string, weight int) *Backend {
	b := Backend{
		addr:   addr,
		weight: weight,
	}
	b.alive.Store(true)

	return &b
}

func (b *Backend) Addr() string {
	return b.addr
}

func (b *Backend) IsAlive() bool {
	return b.alive.Load()
}

func (b *Backend) SetAlive(status bool) {
	b.alive.Store(status)
}

func (b *Backend) Weight() int {
	return b.weight
}

func (b *Backend) Conns() int64 {
	return b.activeConns.Load()
}

func (b *Backend) AddConn() {
	b.activeConns.Add(1)
}

func (b *Backend) RemoveConn() {
	b.activeConns.Add(-1)
}

type Pool struct {
	backends []*Backend
	rotation []*Backend
	counter  atomic.Uint64
}

func NewPool(backends []*Backend) *Pool {
	rotationCapacity := 0

	for i := range backends {
		rotationCapacity = rotationCapacity + backends[i].Weight()
	}

	r := make([]*Backend, 0, rotationCapacity)
	for i := range backends {
		w := backends[i].Weight()
		if w <= 0 {
			w = 1
		}
		for range w {
			r = append(r, backends[i])
		}
	}

	p := Pool{
		backends: backends,
		rotation: r,
	}

	return &p
}

func (p *Pool) Backends() []*Backend {
	return p.backends
}

func (p *Pool) Next() (*Backend, error) {
	length := uint64(len(p.rotation))

	if length == 0 {
		return nil, errors.New("no backends to be found")
	}

	start := (p.counter.Add(1) - 1) % length
	for i := range length {
		idx := (start + i) % length
		b := p.rotation[idx]
		if b.IsAlive() {
			return b, nil
		}
	}

	return nil, errors.New("no healthy backends")
}
