package server

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/mirkobrombin/goup/internal/config"
	"github.com/mirkobrombin/goup/internal/plugin"
	"github.com/mirkobrombin/goup/internal/server/middleware"
	log "github.com/sirupsen/logrus"
)

// createHandler creates the HTTP handler for a site configuration.
func createHandler(conf config.SiteConfig, logger *log.Logger, identifier string, globalMwManager *middleware.MiddlewareManager) (http.Handler, error) {
	var handler http.Handler

	if conf.ProxyPass != "" {
		// Set up reverse proxy handler if ProxyPass is set.
		proxyURL, err := url.Parse(conf.ProxyPass)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL: %v", err)
		}
		proxy := httputil.NewSingleHostReverseProxy(proxyURL)
		handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			addCustomHeaders(w, conf.CustomHeaders)
			proxy.ServeHTTP(w, r)
		})
	} else {
		// Serve static files from the root directory.
		fs := http.FileServer(http.Dir(conf.RootDirectory))
		handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			addCustomHeaders(w, conf.CustomHeaders)
			fs.ServeHTTP(w, r)
		})
	}

	// Copy the global middleware manager for this site
	siteMwManager := globalMwManager.Copy()

	// Initialize plugins for this site
	pluginManager := plugin.GetPluginManagerInstance()
	if err := pluginManager.InitPluginsForSite(conf, logger); err != nil {
		return nil, fmt.Errorf("error initializing plugins for site %s: %v", conf.Domain, err)
	}

	// Add per-site middleware
	reqTimeout := conf.RequestTimeout
	if reqTimeout == 0 {
		reqTimeout = 60 // Default to 60 seconds
	}
	timeout := time.Duration(conf.RequestTimeout) * time.Second
	siteMwManager.Use(middleware.TimeoutMiddleware(timeout))

	// Add logging middleware last to ensure it wraps the entire request
	siteMwManager.Use(middleware.LoggingMiddleware(logger, conf.Domain, identifier))

	// Apply the final chain of middleware
	handler = siteMwManager.Apply(handler)

	return handler, nil
}

// addCustomHeaders adds custom headers to the HTTP response.
func addCustomHeaders(w http.ResponseWriter, headers map[string]string) {
	for key, value := range headers {
		w.Header().Set(key, value)
	}

	// Expose custom headers to the client.
	exposeHeaders := []string{}
	for key := range headers {
		exposeHeaders = append(exposeHeaders, key)
	}

	w.Header().Set("Access-Control-Expose-Headers", strings.Join(exposeHeaders, ", "))
}
