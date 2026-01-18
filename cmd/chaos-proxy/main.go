package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/khizar-sudo/chaos-proxy/internal/chaos"
	"github.com/khizar-sudo/chaos-proxy/internal/config"
	"github.com/khizar-sudo/chaos-proxy/internal/middleware"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(cfg.UpstreamURL)

	chaosConfig := chaos.Config{
		DropRate:    0,
		ErrorRate:   0,
		ErrorCode:   503,
		LatencyMin:  100 * time.Millisecond,
		LatencyMax:  1000 * time.Millisecond,
		CorruptRate: 100,
	}

	chaosEngine := chaos.NewEngine(chaosConfig)

	// Customize the Director to properly set headers for the upstream request
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = cfg.UpstreamURL.Host
		req.Header.Set("User-Agent", "chaos-proxy/1.0")
		req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	}

	handler := middleware.ChaosMiddleware(
		middleware.LoggingMiddleware(
			http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					fmt.Printf("[PROXY] %s %s\n", r.Method, r.URL.Path)
					proxy.ServeHTTP(w, r)
				})),
		chaosEngine)

	fmt.Printf("Starting chaos-proxy on %s\n", cfg.Listen)
	fmt.Printf("Proxying to: %s\n", cfg.UpstreamURL.String())

	fmt.Println("Chaos configuration:")
	fmt.Printf("- Drop rate: %v%%\n", chaosConfig.DropRate)
	if chaosConfig.Latency == 0 {
		fmt.Printf("- Latency injection: %v-%v\n", chaosConfig.LatencyMin, chaosConfig.LatencyMax)
	} else {
		fmt.Printf("- Fixed latency injection: %vms\n", chaosConfig.Latency)
	}
	fmt.Printf("- Error injection: %v%% chance of %d\n", chaosConfig.ErrorRate, chaosConfig.ErrorCode)
	fmt.Printf("- Corruption injection: %v%%\n", chaosConfig.CorruptRate)
	fmt.Println("================================================")

	if err := http.ListenAndServe(cfg.Listen, handler); err != nil {
		log.Fatal(err)
	}
}
