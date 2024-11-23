package middleware

import (
	"net/http"
	"time"

	"github.com/mirkobrombin/goup/internal/tui"
	log "github.com/sirupsen/logrus"
)

// LoggingMiddleware logs HTTP requests.
func LoggingMiddleware(logger *log.Logger, domain string, identifier string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := &responseWriter{ResponseWriter: w, statusCode: 200}

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

// TimeoutMiddleware applies a timeout to HTTP requests.
func TimeoutMiddleware(timeout time.Duration, next http.Handler) http.Handler {
	return http.TimeoutHandler(next, timeout, "Request timed out")
}

// responseWriter is a wrapper for http.ResponseWriter that captures the status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
