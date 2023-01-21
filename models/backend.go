package models

import (
	"log"
	"net"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

type Backend struct {
	URL          *url.URL
	Alive        bool
	mux          sync.RWMutex
	ReverseProxy *httputil.ReverseProxy
}

func (b *Backend) IsAlive() (alive bool) {
	b.mux.RLock()
	alive = b.Alive
	defer b.mux.RUnlock()
	return
}

func (b *Backend) SetAlive(alive bool) {
	b.mux.Lock()
	b.Alive = alive
	defer b.mux.Unlock()
}

func (backend *Backend) isBackendAlive() bool {
	// check the health of backend by pinging it
	timeout := 2 * time.Second
	// create connection with the backend
	conn, err := net.DialTimeout("tcp", backend.URL.Host, timeout)
	if err != nil {
		log.Printf("Backend %s is down: %s", backend.URL.String(), err)
		return false
	}
	// close the connection once we are done with it
	defer conn.Close()
	return true
}
