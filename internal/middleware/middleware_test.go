package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/khizar-sudo/chaos-proxy/internal/chaos"
)

func TestChaosMiddleware_NoChaos(t *testing.T) {
	engine := chaos.NewEngine(chaos.ChaosConfig{})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	middleware := ChaosMiddleware(handler, engine)

	req := httptest.NewRequest("GET", "http://example.com", nil)
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
	if rec.Body.String() != "success" {
		t.Errorf("Expected body 'success', got '%s'", rec.Body.String())
	}
}

func TestChaosMiddleware_Drop(t *testing.T) {
	engine := chaos.NewEngine(chaos.ChaosConfig{
		DropRate: 100,
	})

	handlerCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	middleware := ChaosMiddleware(handler, engine)

	// Create a context with timeout so we can test drop behavior
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	if handlerCalled {
		t.Error("Expected handler to not be called when request is dropped")
	}

	// When dropped, no response is written
	if rec.Code != 200 { // httptest.NewRecorder defaults to 200 if not set
		t.Errorf("Expected no status code to be set, got %d", rec.Code)
	}
}

func TestChaosMiddleware_Error(t *testing.T) {
	errorCode := 503
	engine := chaos.NewEngine(chaos.ChaosConfig{
		ErrorRate: 100,
		ErrorCode: errorCode,
	})

	handlerCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	middleware := ChaosMiddleware(handler, engine)

	req := httptest.NewRequest("GET", "http://example.com", nil)
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	if handlerCalled {
		t.Error("Expected handler to not be called when error is injected")
	}

	if rec.Code != errorCode {
		t.Errorf("Expected status code %d, got %d", errorCode, rec.Code)
	}
}

func TestChaosMiddleware_Latency(t *testing.T) {
	latency := 100 * time.Millisecond
	engine := chaos.NewEngine(chaos.ChaosConfig{
		Latency: latency,
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	middleware := ChaosMiddleware(handler, engine)

	req := httptest.NewRequest("GET", "http://example.com", nil)
	rec := httptest.NewRecorder()

	start := time.Now()
	middleware.ServeHTTP(rec, req)
	duration := time.Since(start)

	// Should take at least the latency duration
	if duration < latency {
		t.Errorf("Expected request to take at least %v, took %v", latency, duration)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
	if rec.Body.String() != "success" {
		t.Errorf("Expected body 'success', got '%s'", rec.Body.String())
	}
}

func TestChaosMiddleware_LatencyCancellation(t *testing.T) {
	latency := 5 * time.Second
	engine := chaos.NewEngine(chaos.ChaosConfig{
		Latency: latency,
	})

	handlerCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	middleware := ChaosMiddleware(handler, engine)

	// Create a context that cancels quickly
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	if handlerCalled {
		t.Error("Expected handler to not be called when request is cancelled during latency")
	}
}

func TestChaosMiddleware_Corrupt(t *testing.T) {
	engine := chaos.NewEngine(chaos.ChaosConfig{
		CorruptRate: 100,
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	middleware := ChaosMiddleware(handler, engine)

	req := httptest.NewRequest("GET", "http://example.com", nil)
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	// Response is corrupted, so it should differ from "success"
	// Note: Due to randomness, we can't predict exact output, but we can verify something was written
	if rec.Body.Len() == 0 {
		t.Error("Expected some response body")
	}
}

func TestChaosMiddleware_ErrorWithLatency(t *testing.T) {
	latency := 100 * time.Millisecond
	errorCode := 503
	engine := chaos.NewEngine(chaos.ChaosConfig{
		ErrorRate: 100,
		ErrorCode: errorCode,
		Latency:   latency,
	})

	handlerCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	middleware := ChaosMiddleware(handler, engine)

	req := httptest.NewRequest("GET", "http://example.com", nil)
	rec := httptest.NewRecorder()

	start := time.Now()
	middleware.ServeHTTP(rec, req)
	duration := time.Since(start)

	// Should take at least the latency duration
	if duration < latency {
		t.Errorf("Expected request to take at least %v, took %v", latency, duration)
	}

	if handlerCalled {
		t.Error("Expected handler to not be called when error is injected")
	}

	if rec.Code != errorCode {
		t.Errorf("Expected status code %d, got %d", errorCode, rec.Code)
	}
}
