package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mirkobrombin/goup/internal/logger"
)

func TestLoggingMiddleware(t *testing.T) {
	testLogger, err := logger.NewLogger("test_middleware_logging", nil)
	if err != nil {
		t.Fatalf("Error creating logger: %v", err)
	}
	testLogger.SetOutput(httptest.NewRecorder())

	identifier := "test_identifier"
	domain := "test_domain"

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Apply middleware to the handler
	loggingMiddleware := LoggingMiddleware(testLogger, domain, identifier)
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

	expectedBody := "Request timed out"
	if w.Body.String() != expectedBody {
		t.Errorf("Expected body %q, got %q", expectedBody, w.Body.String())
	}
}
