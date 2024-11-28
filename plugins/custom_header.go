package plugins

import (
	"net/http"

	"github.com/mirkobrombin/goup/internal/config"
	"github.com/mirkobrombin/goup/internal/server/middleware"
	log "github.com/sirupsen/logrus"
)

// CustomHeaderPlugin adds a custom header to all HTTP responses.
type CustomHeaderPlugin struct{}

// Name returns the name of the plugin.
func (p *CustomHeaderPlugin) Name() string {
	return "CustomHeaderPlugin"
}

// Init registers any global middleware (none for CustomHeaderPlugin).
func (p *CustomHeaderPlugin) Init(mwManager *middleware.MiddlewareManager) error {
	return nil
}

// InitForSite initializes the plugin for a specific site.
func (p *CustomHeaderPlugin) InitForSite(mwManager *middleware.MiddlewareManager, logger *log.Logger, conf config.SiteConfig) error {
	mwManager.Use(p.customHeaderMiddleware(logger, conf))
	return nil
}

// customHeaderMiddleware is the middleware that adds custom headers.
func (p *CustomHeaderPlugin) customHeaderMiddleware(logger *log.Logger, conf config.SiteConfig) middleware.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for key, value := range conf.CustomHeaders {
				w.Header().Set(key, value)
				logger.Infof("Added custom header %s: %s", key, value)
			}
			next.ServeHTTP(w, r)
		})
	}
}
