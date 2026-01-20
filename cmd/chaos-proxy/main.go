package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/khizar-sudo/chaos-proxy/internal/chaos"
	"github.com/khizar-sudo/chaos-proxy/internal/config"
	"github.com/khizar-sudo/chaos-proxy/internal/middleware"
	"github.com/khizar-sudo/chaos-proxy/internal/watcher"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	configPath := "config.yaml"
	watcher, err := watcher.NewWatcher(configPath)
	if err != nil {
		slog.Warn("failed to set up config watcher, hot reload disabled")
	} else {
		watcher.Start()
		defer watcher.Close()
		slog.Info("config file watching enabled", "path", configPath)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	for {
		srv, err := startServer(cfg)
		if err != nil {
			return err
		}

		select {
		case <-sigChan:
			slog.Info("shutdown signal received, stopping server")
			shutDownServer(srv)
			return nil
		case <-watcher.ReloadChan():
			slog.Info("reloading configuration...")
			shutDownServer(srv)

			newCfg, err := config.Load()
			if err != nil {
				slog.Error("failed to reload config", "error", err)
				slog.Info("keeping previous configuration")
			} else {
				cfg = newCfg
				slog.Info("configuration reloaded successfully")
			}
		}
	}
}

func startServer(cfg *config.Config) (*http.Server, error) {
	proxy := httputil.NewSingleHostReverseProxy(cfg.UpstreamURL)
	latencies, err := cfg.ParseDurations()
	if err != nil {
		return nil, err
	}

	chaosConfig := chaos.ChaosConfig{
		DropRate:    cfg.Chaos.DropRate,
		ErrorRate:   cfg.Chaos.ErrorRate,
		ErrorCode:   cfg.Chaos.ErrorCode,
		Latency:     latencies.Latency,
		LatencyMin:  latencies.LatencyMin,
		LatencyMax:  latencies.LatencyMax,
		CorruptRate: cfg.Chaos.CorruptRate,
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
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Printf("[PROXY] %s %s\n", r.Method, r.URL.Path)
				proxy.ServeHTTP(w, r)
			}),
		chaosEngine)
	handler = middleware.LoggingMiddleware(handler)

	srv := &http.Server{
		Addr:              cfg.Listen,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		fmt.Println()
		fmt.Println("==============================================================================")
		slog.Info("starting server", "listen", cfg.Listen, "upstream", cfg.UpstreamURL.String())
		cfg.PrintConfiguration()
		if err := srv.ListenAndServe(); err != nil {
			slog.Error("server error", "error", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	return srv, nil
}

func shutDownServer(srv *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
		fmt.Println()
	}()

	slog.Info("shutting down server...")
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server shutdown error", "error", err)
	} else {
		slog.Info("server stopped gracefully")
	}
}
