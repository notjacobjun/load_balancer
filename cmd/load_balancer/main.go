package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	lb "github.com/notjacobjun/load_balancer/loadbalancer"
	models "github.com/notjacobjun/load_balancer/models"
	s "github.com/notjacobjun/load_balancer/server"
)

var (
	Server http.Server
)

func healthCheck() {
	// start the timer that runs every 20 seconds
	t := time.NewTicker(20 * time.Second)

	for {
		select {
		case <-t.C:
			log.Printf("Checking the health of the backends")
			// check the health of the backends
			models.Pool.HealthCheck()
			log.Printf("Health check complete")
		}
	}
}

func main() {
	var serverList string
	var port int
	flag.StringVar(&serverList, "backends", "", "Load balanced backends, use commas to separate")
	flag.IntVar(&port, "port", 3030, "Port to serve")
	flag.Parse()

	if len(serverList) == 0 {
		log.Fatal("Please provide one or more backends to load balance")
	}

	tokens := strings.Split(serverList, ",")
	for _, tok := range tokens {
		u, err := url.Parse(tok)
		if err != nil {
			log.Fatal(err)
		}

		// initialize the reverse proxy
		rp := httputil.NewSingleHostReverseProxy(u)
		// setup the error handler for this reverse proxy
		rp.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			// log the current request with the error
			log.Printf("[%s] %s\n", u.Host, err.Error())
			// get the retries from the context
			retries := s.GetRetryFromContext(r)
			if retries < 3 {
				select {
				// if we have waited for 10 milliseconds, retry the request
				case <-time.After(10 * time.Millisecond):
					// increment the number of retries attached to this context
					ctx := context.WithValue(r.Context(), s.Retry, retries+1)
					// perform the retry
					rp.ServeHTTP(w, r.WithContext(ctx))
				}
				return
			}

			// if we have exceeded the number of retries, send the traffic to the next backend
			// mark this backend as unhealthy
			models.Pool.MarkBackendStatus(u, false)

			// get the number of attempts from the http context
			attempts := s.GetAttemptsFromContext(r)
			// log our new attempt
			log.Printf("%s(%s) Attempting retry %d\n", r.RemoteAddr, r.URL.Path, attempts)
			// update the attempts value
			ctx := context.WithValue(r.Context(), s.Attempts, attempts+1)
			lb.LoadBalance(w, r.WithContext(ctx))
		}

		// add this new backend to the pool
		models.Pool.AddBackend(&models.Backend{
			URL: u,
			// set the status to true
			Alive: true,
			// set the reverse proxy
			ReverseProxy: rp,
		})
		log.Printf("Added backend %s to the pool", u)
	}
	// setup the server
	Server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: http.HandlerFunc(lb.LoadBalance),
	}
	// start the healthcheck goroutine
	go healthCheck()
	// start the load balancer
	log.Printf("Load Balancer started at :%d\n", port)
	if err := Server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
