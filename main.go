package main

import (
    "fmt"
    "net/http"
    "net/http/httputil"
    "net/url"
    "os"
)

type Server interface {
    Address() string
    IsAlive() bool
    Serve(w http.ResponseWriter, r *http.Request)
}

// LoadBalancer struct
type LoadBalancer struct {
    port             string
    roundRobinCount  int
    servers          []Server
}

func newLoadBalancer(port string, servers []Server) *LoadBalancer {
    return &LoadBalancer{
        port:            port,
        roundRobinCount: 0,
        servers:         servers,
    }
}

// SimpleServer struct
type SimpleServer struct {
    addr  string
    proxy *httputil.ReverseProxy
}

func newServer(addr string) *SimpleServer {
    serveUrl, err := url.Parse(addr)
    handleErr(err)
    return &SimpleServer{
        addr:  addr,
        proxy: httputil.NewSingleHostReverseProxy(serveUrl),
    }
}

func handleErr(err error) {
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }
}

func (s *SimpleServer) Address() string {
    return s.addr
}

func (s *SimpleServer) IsAlive() bool {
    return true
}

func (s *SimpleServer) Serve(w http.ResponseWriter, r *http.Request) {
    s.proxy.ServeHTTP(w, r)
}

func (lb *LoadBalancer) getNextAvailabeServer() Server {
    server := lb.servers[lb.roundRobinCount % len(lb.servers)]
    for !server.IsAlive() {
        lb.roundRobinCount++
        server = lb.servers[lb.roundRobinCount % len(lb.servers)]
    }
    lb.roundRobinCount++
    return server
}

func (lb *LoadBalancer) serveProxy(w http.ResponseWriter, r *http.Request) {
    targetServer := lb.getNextAvailabeServer()
    fmt.Printf("forwarding requests to address %q\n", targetServer.Address())
    targetServer.Serve(w, r)
}

func main() {
    servers := []Server{
        newServer("https://www.youtube.com/"),
        newServer("https://www.reddit.com/"),
        newServer("https://leetcode.com/"),
    }
    lb := newLoadBalancer("8080", servers)
    handleRedirect := func(w http.ResponseWriter, r *http.Request) {
        lb.serveProxy(w, r)
    }
    http.HandleFunc("/", handleRedirect)

    fmt.Printf("serving requests at localhost:%s\n", lb.port)
    err := http.ListenAndServe(":"+lb.port, nil)
    if err != nil {
        handleErr(err)
    }
}

