package lb

import (
	"io"
	"log"
	"net"
	"net/url"

	"github.com/PraneethV-cmd/frost-load/internal/pool"
)

type LoadBalancer4 struct {
	pool *pool.Pool
}

func NewLoadbalancer4(pool *pool.Pool) *LoadBalancer4 {
	lb4 := LoadBalancer4{
		pool: pool,
	}
	return &lb4
}

func (lb4 *LoadBalancer4) ListenAndHandle(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go lb4.handleConnection(conn)
	}
}

func (lb4 *LoadBalancer4) handleConnection(conn net.Conn) {
	defer conn.Close()

	target, err := lb4.pool.Next()
	if err != nil {
		return
	}

	u, err := url.Parse(target.Addr())
	if err != nil {
		return
	}

	upstream, err := net.Dial("tcp", u.Host)
	if err != nil {
		target.SetAlive(false)
		return
	}

	defer upstream.Close()

	target.AddConn()
	defer target.RemoveConn()

	fullDuplexPipes := make(chan struct{}, 2)

	go func() {
		io.Copy(upstream, conn)
		fullDuplexPipes <- struct{}{}
	}()

	go func() {
		io.Copy(conn, upstream)
		fullDuplexPipes <- struct{}{}
	}()

	<-fullDuplexPipes
}
