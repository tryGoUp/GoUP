package middleware

import (
	"net/http"
	"time"

	"github.com/mirkobrombin/goup/internal/tui"
	log "github.com/sirupsen/logrus"
)

// LoggingMiddleware logs HTTP requests.
func LoggingMiddleware(logger *log.Logger, domain string, identifier string) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(rw, r)

			duration := time.Since(start)

			entry := logger.WithFields(log.Fields{
				"domain":      domain,
				"identifier":  identifier,
				"method":      r.Method,
				"url":         r.URL.String(),
				"remote_addr": r.RemoteAddr,
				"status_code": rw.statusCode,
				"duration":    duration.Seconds(),
			})
			entry.Info("Handled request")

			if tui.IsEnabled() {
				tui.UpdateLog(identifier, entry)
			}
		})
	}
}

// TimeoutMiddleware applies a timeout to HTTP requests.
func TimeoutMiddleware(timeout time.Duration) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.TimeoutHandler(next, timeout, "Request timed out")
	}
}

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader sets the HTTP status code.
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
