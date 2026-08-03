package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PraneethV-cmd/frost-load/internal/lb"
	"github.com/PraneethV-cmd/frost-load/internal/pool"
)

func main() {
	//	b1 := pool.NewBackend("http://localhost:8080", 1)
	//	b2 := pool.NewBackend("http://localhost:8081", 1)
	//	b3 := pool.NewBackend("http://localhost:8082", 3)
	//
	//backends := []*pool.Backend{b1, b2, b3}

	raw := os.Getenv("BACKENDS")
	if raw == "" {
		raw = "http://localhost:8080, http://localhost:8081, http://localhost:8082"
	}

	var backends []*pool.Backend
	for _, b := range strings.Split(raw, ",") {
		bAddr := strings.TrimSpace(b)
		if bAddr == "" {
			continue
		}
		backends = append(backends, pool.NewBackend(bAddr, 1))
	}

	if len(backends) == 0 {
		log.Fatal("no backends configured")
	}

	p1 := pool.NewPool(backends, pool.LeastConns{})
	// p2 := pool.NewPool(backends, pool.RoundRobin{})

	balancer, err := lb.New(p1)
	l4 := lb.NewLoadbalancer4(p1)

	if err != nil {
		log.Fatal(err)
		return
	}

	s := &http.Server{
		Addr:              ":9000",
		Handler:           balancer,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("frostload running on 9000")

	go p1.HealthCheck(3 * time.Second)

	go func() {
		log.Fatal(l4.ListenAndHandle(":9090"))
	}()
	log.Fatal(s.ListenAndServe())
}
