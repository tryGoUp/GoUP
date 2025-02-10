package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/mirkobrombin/goup/internal/logger"
	"github.com/mirkobrombin/goup/internal/tui"
)

// LoggingMiddleware logs HTTP requests.
func LoggingMiddleware(l *logger.Logger, domain string, identifier string) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(rw, r)

			duration := time.Since(start)

			// Extract the real IP address if behind proxies
			remoteAddr := r.RemoteAddr
			if ip := r.Header.Get("X-Real-IP"); ip != "" {
				remoteAddr = ip
			} else if ips := r.Header.Get("X-Forwarded-For"); ips != "" {
				remoteAddr = ips
			}

			fields := logger.Fields{
				"method":       r.Method,
				"url":          r.URL.String(),
				"remote_addr":  remoteAddr,
				"status_code":  rw.statusCode,
				"duration_sec": duration.Seconds(),
				"domain":       domain,
			}
			l.WithFields(fields).Info("Handled request")

			if tui.IsEnabled() {
				tui.UpdateLog(identifier, fields)
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

// BenchmarkMiddleware logs the duration of HTTP requests.
func BenchmarkMiddleware() MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			next.ServeHTTP(w, r)

			duration := time.Since(start)
			fmt.Printf("\033[33;40m⏲ Benchmark: %s %s completed in %s\033[0m\n",
				r.Method, r.URL.Path, formatDuration(duration))
		})
	}
}

// formatDuration formats a time.Duration to a human-readable string.
func formatDuration(d time.Duration) string {
	switch {
	case d < time.Millisecond:
		return fmt.Sprintf("%dµs", d.Microseconds())
	case d < time.Second:
		return fmt.Sprintf("%.3fms", float64(d.Microseconds())/1000)
	default:
		return fmt.Sprintf("%.3fs", d.Seconds())
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
