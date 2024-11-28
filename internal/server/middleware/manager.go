package middleware

import "net/http"

// MiddlewareFunc represents the type for middleware functions.
type MiddlewareFunc func(http.Handler) http.Handler

// MiddlewareManager manages a chain of middleware.
type MiddlewareManager struct {
	middleware []MiddlewareFunc
}

// NewMiddlewareManager creates a new instance of MiddlewareManager.
func NewMiddlewareManager() *MiddlewareManager {
	return &MiddlewareManager{
		middleware: []MiddlewareFunc{},
	}
}

// Use adds a middleware function to the chain.
func (m *MiddlewareManager) Use(mw MiddlewareFunc) {
	m.middleware = append(m.middleware, mw)
}

// Apply applies all registered middleware to a handler.
func (m *MiddlewareManager) Apply(handler http.Handler) http.Handler {
	for i := len(m.middleware) - 1; i >= 0; i-- {
		handler = m.middleware[i](handler)
	}
	return handler
}

// Copy creates a new MiddlewareManager with the same middleware chain.
func (m *MiddlewareManager) Copy() *MiddlewareManager {
	copiedMiddleware := make([]MiddlewareFunc, len(m.middleware))
	copy(copiedMiddleware, m.middleware)
	return &MiddlewareManager{
		middleware: copiedMiddleware,
	}
}
