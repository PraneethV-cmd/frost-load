package lb

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/PraneethV-cmd/frost-load/internal/pool"
)

type LoadBalancer struct {
	pool    *pool.Pool
	proxies map[*pool.Backend]*httputil.ReverseProxy
}

// constructor

func New(p *pool.Pool) (*LoadBalancer, error) {
	proxy := make(map[*pool.Backend]*httputil.ReverseProxy)

	for _, b := range p.Backends() {
		u, err := url.Parse(b.Addr())
		if err != nil {
			return nil, err
		}
		proxy[b] = httputil.NewSingleHostReverseProxy(u)
	}

	return &LoadBalancer{
		pool:    p,
		proxies: proxy,
	}, nil
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b, err := lb.pool.Next()
	if err != nil {
		http.Error(w, "no healthy backends", http.StatusServiceUnavailable)
		return
	}

	b.AddConn()
	defer b.RemoveConn()
	lb.proxies[b].ServeHTTP(w, r)
}

