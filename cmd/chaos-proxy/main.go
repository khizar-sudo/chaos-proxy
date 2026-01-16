package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"

	"github.com/khizar-sudo/chaos-proxy/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(cfg.UpstreamURL)

	// Customize the Director to properly set headers for the upstream request (problems with Cloudflare otherwise)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = cfg.UpstreamURL.Host
		req.Header.Set("User-Agent", "chaos-proxy/1.0")
		req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("[PROXY] %s %s\n", r.Method, r.URL.Path)
		proxy.ServeHTTP(w, r)
	})

	fmt.Printf("Starting chaos-proxy on %s\n", cfg.Listen)
	fmt.Printf("Proxying to: %s\n", cfg.UpstreamURL.String())

	if err := http.ListenAndServe(cfg.Listen, nil); err != nil {
		log.Fatal(err)
	}
}
