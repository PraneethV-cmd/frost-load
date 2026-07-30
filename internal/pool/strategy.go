package pool

import "errors"

type Strategy interface {
	Pick(p *Pool) (*Backend, error)
}

type (
	RoundRobin struct{}
	LeastConns struct{}
)

func (RoundRobin) Pick(p *Pool) (*Backend, error) {
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

func (LeastConns) Pick(p *Pool) (*Backend, error) {
	var leastConnBackend *Backend
	for _, b := range p.backends {
		if !b.IsAlive() {
			continue
		}

		if leastConnBackend == nil || b.Conns() < leastConnBackend.Conns() {
			leastConnBackend = b
		}
	}
	if leastConnBackend == nil {
		return nil, errors.New("no healthy backends")
	}
	return leastConnBackend, nil
}
