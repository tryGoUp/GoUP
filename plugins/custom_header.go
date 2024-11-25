package plugins

import (
	"net/http"

	"github.com/mirkobrombin/goup/internal/server/middleware"
)

// CustomHeaderPlugin adds a custom header to all HTTP responses.
type CustomHeaderPlugin struct{}

// Name returns the name of the plugin.
func (p *CustomHeaderPlugin) Name() string {
	return "CustomHeaderPlugin"
}

// Init initializes the plugin and registers its middleware.
func (p *CustomHeaderPlugin) Init(mwManager *middleware.MiddlewareManager) error {
	mwManager.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-GoUP-Header", "GoUP")
			next.ServeHTTP(w, r)
		})
	})
	return nil
}
