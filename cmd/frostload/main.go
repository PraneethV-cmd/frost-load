package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/PraneethV-cmd/frost-load/internal/lb"
	"github.com/PraneethV-cmd/frost-load/internal/pool"
)

func main() {
	b1 := pool.NewBackend("http://localhost:8080", 1)
	b2 := pool.NewBackend("http://localhost:8081", 1)
	b3 := pool.NewBackend("http://localhost:8082", 3)

	backends := []*pool.Backend{b1, b2, b3}

	p1 := pool.NewPool(backends)

	balancer, err := lb.New(p1)
	if err != nil {
		log.Fatal(err)
		return
	}

	s := &http.Server{
		Addr:              ":9000",
		Handler:           balancer,
		ReadHeaderTimeout: 5 * time.Second,
	}

	fmt.Println("frostload on :9000")
	log.Fatal(s.ListenAndServe())
}
