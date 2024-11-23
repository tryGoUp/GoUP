package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
)

func TestLoggingMiddleware(t *testing.T) {
	logger := log.New()
	logger.Out = httptest.NewRecorder() // Discard output on purpose

	domain := "example.com"
	identifier := "test"

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})

	middlewareHandler := LoggingMiddleware(logger, domain, identifier, handler)

	req := httptest.NewRequest("GET", "http://example.com/test", nil)
	w := httptest.NewRecorder()

	middlewareHandler.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status code 200, got %d", w.Code)
	}
}

func TestTimeoutMiddleware(t *testing.T) {
	timeout := 1 * time.Second

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.Write([]byte("This should timeout"))
	})

	middlewareHandler := TimeoutMiddleware(timeout, handler)

	req := httptest.NewRequest("GET", "http://example.com/test", nil)
	w := httptest.NewRecorder()

	middlewareHandler.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status code %d, got %d", http.StatusServiceUnavailable, w.Code)
	}

	expectedBody := "Request timed out"
	if w.Body.String() != expectedBody {
		t.Errorf("Expected body %q, got %q", expectedBody, w.Body.String())
	}
}
