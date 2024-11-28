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
	logger.Out = httptest.NewRecorder() // Discard output for testing

	identifier := "test_identifier"

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Get the middleware function
	loggingMiddleware := LoggingMiddleware(logger, "test_domain", identifier)

	// Apply middleware to the handler
	middlewareHandler := loggingMiddleware(handler)

	req := httptest.NewRequest("GET", "http://example.com/test", nil)
	w := httptest.NewRecorder()

	middlewareHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestTimeoutMiddleware(t *testing.T) {
	timeout := 1 * time.Second

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.Write([]byte("This should timeout"))
	})

	// Get the middleware function
	timeoutMiddleware := TimeoutMiddleware(timeout)

	// Apply middleware to the handler
	middlewareHandler := timeoutMiddleware(handler)

	req := httptest.NewRequest("GET", "http://example.com/test", nil)
	w := httptest.NewRecorder()

	middlewareHandler.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status code %d, got %d", http.StatusServiceUnavailable, w.Code)
	}

	expectedBody := "Request timed out\n"
	if w.Body.String() != expectedBody {
		t.Errorf("Expected body %q, got %q", expectedBody, w.Body.String())
	}
}
