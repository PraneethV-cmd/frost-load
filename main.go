package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
)

//So basically the load balancer that I will be making is of a round-robin variety
//SO in case of some failure in the backend , we will have to route the traffic to the
// backend thats up and running

//this is a simple backend frame or overall strcuture of the backedn
// AliveStatus is for seeing whrther a paritcular backend is dead or not
// reverseproxy is an http handler that take incoming request
// and then we send it to another server thereby proxying the response back to client

type Backend struct {
    URL *url.URL
    AliveStatus bool
    mux sync.RWMutex
    ReverseProxy *httputil.ReverseProxy
}


//now to keep track of all the backends in our load balancer ,
// we use a slice  

type ServerPool struct {
    backends []*Backend
    current uint64 
}

// to skip the dead backedsn and then pick the proper one we need to keep a count of the 
// backends

//here we are increasing the current valye bu one atomically 
// and then we reutnr the index by modding it with the length of the slice
// so the value will always be between o and the length of the sluice
// in the end, we are interested only in the index 

func (s *ServerPool) NextIndex() int {
    return int(atomic.AddUint64(&s.current, uint64(1)) % uint64(len(s.backends)))
}

// now to take the next backend whilst making sure to skip the dead ones ... 

func (s *ServerPool) GetNextPeer() *Backend {
    next := s.NextIndex()
    l := len(s.backends) + next  //to start from the next and then do an entire cycle 
    for i := next; i < l ; i++ {
        idx := i % len(s.backends)
        if s.backends[idx].IsAlive() {
            if i != next {
                atomic.StoreUint64(&s.current, uint64(idx)) // to mark the current one 
            }
            return s.backends[idx]
        }
    }
    return nil 
}

func (b *Backend) SetAlive(alive bool){
    b.mux.Lock()
    b.AliveStatus = alive
    b.mux.Unlock()
}

func (b *Backend) IsAlive() (alive bool){
    b.mux.RLock() //acquire a read lock 
    alive = b.AliveStatus  
    b.mux.RUnlock() //release the read lock that we got
    return 
}

func lb(w http.ResponseWriter, r *http.Request) {
    peer := ServerPool.GetNextPeer() 
    if peer != nil {
        peer.ReverseProxy.ServeHTTP(w,r)
        return 
    }
    http.Error(w, "Service not available", http.StatusServiceUnavailable)
}

