package server

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/mirkobrombin/goup/internal/config"
	"github.com/mirkobrombin/goup/internal/server/middleware"
	log "github.com/sirupsen/logrus"
)

// createHandler creates the HTTP handler for a site configuration.
func createHandler(conf config.SiteConfig, logger *log.Logger, identifier string) (http.Handler, error) {
	var handler http.Handler

	if conf.ProxyPass != "" {
		// Set up reverse proxy handler if ProxyPass is set
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
		// Serve static files
		fs := http.FileServer(http.Dir(conf.RootDirectory))
		handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			addCustomHeaders(w, conf.CustomHeaders)
			fs.ServeHTTP(w, r)
		})
	}

	// Wrap with logging and timeout middleware
	timeout := time.Duration(conf.RequestTimeout) * time.Second
	handler = middleware.TimeoutMiddleware(timeout, middleware.LoggingMiddleware(logger, conf.Domain, identifier, handler))

	return handler, nil
}

// addCustomHeaders adds custom headers to the HTTP response.
func addCustomHeaders(w http.ResponseWriter, headers map[string]string) {
	for key, value := range headers {
		w.Header().Set(key, value)
	}

	// Expose the custom headers to the client. Should be safe (?)
	exposeHeaders := []string{}
	for key := range headers {
		exposeHeaders = append(exposeHeaders, key)
	}

	w.Header().Set("Access-Control-Expose-Headers", strings.Join(exposeHeaders, ", "))
}
