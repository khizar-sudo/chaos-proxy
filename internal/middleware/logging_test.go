package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoggingMiddleware_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	middleware := LoggingMiddleware(handler)

	req := httptest.NewRequest("GET", "http://example.com/test", nil)
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
	if rec.Body.String() != "success" {
		t.Errorf("Expected body 'success', got '%s'", rec.Body.String())
	}
}

func TestLoggingMiddleware_Error(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error"))
	})

	middleware := LoggingMiddleware(handler)

	req := httptest.NewRequest("POST", "http://example.com/api/data", nil)
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", rec.Code)
	}
}

func TestLoggingMiddleware_DifferentMethods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			middleware := LoggingMiddleware(handler)

			req := httptest.NewRequest(method, "http://example.com/test", nil)
			rec := httptest.NewRecorder()

			middleware.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", rec.Code)
			}
		})
	}
}

func TestStatusRecorder_WriteHeader(t *testing.T) {
	rec := httptest.NewRecorder()
	sr := &statusRecorder{
		ResponseWriter: rec,
		statusCode:     http.StatusOK,
	}

	sr.WriteHeader(http.StatusNotFound)

	if sr.statusCode != http.StatusNotFound {
		t.Errorf("Expected status code to be 404, got %d", sr.statusCode)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected underlying recorder status to be 404, got %d", rec.Code)
	}
}

func TestStatusRecorder_Write(t *testing.T) {
	rec := httptest.NewRecorder()
	sr := &statusRecorder{
		ResponseWriter: rec,
		statusCode:     http.StatusOK,
		written:        0,
	}

	data := []byte("test data")
	n, err := sr.Write(data)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if n != len(data) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(data), n)
	}
	if sr.written != len(data) {
		t.Errorf("Expected written counter to be %d, got %d", len(data), sr.written)
	}
	if rec.Body.String() != string(data) {
		t.Errorf("Expected body to be '%s', got '%s'", string(data), rec.Body.String())
	}
}

func TestStatusRecorder_MultipleWrites(t *testing.T) {
	rec := httptest.NewRecorder()
	sr := &statusRecorder{
		ResponseWriter: rec,
		statusCode:     http.StatusOK,
		written:        0,
	}

	data1 := []byte("first ")
	data2 := []byte("second")

	n1, err1 := sr.Write(data1)
	n2, err2 := sr.Write(data2)

	if err1 != nil || err2 != nil {
		t.Errorf("Expected no errors, got %v, %v", err1, err2)
	}

	totalWritten := n1 + n2
	if sr.written != totalWritten {
		t.Errorf("Expected written counter to be %d, got %d", totalWritten, sr.written)
	}
	if rec.Body.String() != "first second" {
		t.Errorf("Expected body to be 'first second', got '%s'", rec.Body.String())
	}
}

func TestLoggingMiddleware_DefaultStatusCode(t *testing.T) {
	// Handler that doesn't explicitly set status code
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("success"))
	})

	middleware := LoggingMiddleware(handler)

	req := httptest.NewRequest("GET", "http://example.com/test", nil)
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	// Should default to 200
	if rec.Code != http.StatusOK {
		t.Errorf("Expected default status 200, got %d", rec.Code)
	}
}

func TestLoggingMiddleware_BytesWritten(t *testing.T) {
	expectedData := "This is a longer response body for testing byte counting"

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(expectedData))
	})

	middleware := LoggingMiddleware(handler)

	req := httptest.NewRequest("GET", "http://example.com/test", nil)
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	if rec.Body.Len() != len(expectedData) {
		t.Errorf("Expected %d bytes written, got %d", len(expectedData), rec.Body.Len())
	}
}
