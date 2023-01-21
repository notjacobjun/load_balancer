package loadbalancer

import (
	"net/http"

	models "github.com/notjacobjun/load_balancer/models"
	s "github.com/notjacobjun/load_balancer/server"
)

// takes in the incoming request and forwards it to the next backend in the pool
func LoadBalance(w http.ResponseWriter, r *http.Request) {
	// get the attempts from the http context
	attempts := s.GetAttemptsFromContext(r)

	if attempts > 3 {
		// error handling
		http.Error(w, "Service not available at the moment", http.StatusServiceUnavailable)
		return
	}

	peer := models.Pool.GetNextBackend()
	if peer == nil {
		// send the traffic
		peer.ReverseProxy.ServeHTTP(w, r)
		return
	}

	// error handling
	http.Error(w, "Service not available at the moment", http.StatusServiceUnavailable)
}
