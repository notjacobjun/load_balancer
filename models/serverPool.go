package models

import (
	"log"
	"net/url"
	"sync/atomic"
)

type ServerPool struct {
	backends []*Backend
	current  uint64
}

var Pool ServerPool

func (p *ServerPool) NextIndex() int {
	// using atomic 'AddUint64' to increment the counter within bounds
	return int(atomic.AddUint64(&p.current, 1) % uint64(len(p.backends)))
}

func (p *ServerPool) GetNextBackend() *Backend {
	// get the next backend based on the current index
	// also based on whether it is alive or not
	next := p.NextIndex()
	// looping through the entire list of backends in the worst case
	for i := next; i < next+len(p.backends); i++ {
		idx := i % len(p.backends)
		// check the health of the backend
		if p.backends[idx].IsAlive() {
			// update the current index if we chose a backend that was different
			// from the one we started with
			if idx != next {
				atomic.StoreUint64(&p.current, uint64(idx))
			}
			return p.backends[idx]
		}
	}
	return nil
}

func (p *ServerPool) AddBackend(backend *Backend) {
	p.backends = append(p.backends, backend)
}

func (p *ServerPool) RemoveBackend(backend *Backend) {
	for i := 0; i < len(p.backends); i++ {
		if p.backends[i] == backend {
			p.backends = append(p.backends[:i], p.backends[i+1:]...)
			return
		}
	}
	log.Default().Printf("Backend %s not found", backend.URL.String())
}

func (p *ServerPool) MarkBackendStatus(hostUrl *url.URL, alive bool) {
	for i := 0; i < len(p.backends); i++ {
		if p.backends[i].URL == hostUrl {
			p.backends[i].SetAlive(alive)
			return
		}
	}
}

func (p *ServerPool) HealthCheck() {
	status := "up"
	for _, backend := range p.backends {
		go func(b *Backend) {
			alive := b.isBackendAlive()
			if !alive {
				status = "down"
			}
			// log the status of this backend
			log.Printf("Backend %s is %s", b.URL.String(), status)
			p.MarkBackendStatus(b.URL, alive)
		}(backend)
	}
}
