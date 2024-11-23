package server

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/mirkobrombin/goup/internal/config"
	log "github.com/sirupsen/logrus"
)

// createHTTPServer creates an HTTP server with the given configuration and handler.
func createHTTPServer(conf config.SiteConfig, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:         fmt.Sprintf(":%d", conf.Port),
		Handler:      handler,
		ReadTimeout:  time.Duration(conf.RequestTimeout) * time.Second,
		WriteTimeout: time.Duration(conf.RequestTimeout) * time.Second,
	}
}

// startServerInstance starts the HTTP server instance.
func startServerInstance(server *http.Server, conf config.SiteConfig, logger *log.Logger) {
	go func() {
		if conf.SSL.Enabled {
			// SSL/TLS configuration
			if _, err := os.Stat(conf.SSL.Certificate); os.IsNotExist(err) {
				logger.Errorf("SSL certificate not found for %s: %v", conf.Domain, err)
				return
			}
			if _, err := os.Stat(conf.SSL.Key); os.IsNotExist(err) {
				logger.Errorf("SSL key not found for %s: %v", conf.Domain, err)
				return
			}

			cert, err := tls.LoadX509KeyPair(conf.SSL.Certificate, conf.SSL.Key)
			if err != nil {
				logger.Errorf("Error loading SSL certificates for %s: %v", conf.Domain, err)
				return
			}

			server.TLSConfig = &tls.Config{Certificates: []tls.Certificate{cert}}

			logger.Infof("Serving %s on HTTPS port %d", conf.Domain, conf.Port)
			if err := server.ListenAndServeTLS("", ""); err != nil {
				logger.Errorf("Server error for %s: %v", conf.Domain, err)
			}
		} else {
			logger.Infof("Serving on HTTP port %d", conf.Port)
			if err := server.ListenAndServe(); err != nil {
				logger.Errorf("Server error on port %d: %v", conf.Port, err)
			}
		}
	}()
}
