package pool

import (
	"net/http"
	"sync"
	"time"
)

var probeClient = &http.Client{
	Timeout: 2 * time.Second,
}

func (p *Pool) CheckHealthSingleProbe(b *Backend) {
	resp, err := probeClient.Get(b.Addr() + "/health")
	if err != nil {
		b.SetAlive(false)
		return
	}

	defer resp.Body.Close()
	b.SetAlive(http.StatusOK == resp.StatusCode)
}

func (p *Pool) HealthCheck(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		var wg sync.WaitGroup
		for _, b := range p.Backends() {
			wg.Add(1)
			go func(b *Backend) {
				defer wg.Done()
				p.CheckHealthSingleProbe(b)
			}(b)
		}
		wg.Wait()
	}
}
