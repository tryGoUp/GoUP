package server

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/mirkobrombin/goup/internal/config"
	"github.com/mirkobrombin/goup/internal/logger"
	"github.com/quic-go/quic-go/http3"
)

// createHTTPServer creates an HTTP server with the given configuration and handler.
func createHTTPServer(conf config.SiteConfig, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:         fmt.Sprintf(":%d", conf.Port),
		Handler:      handler,
		ReadTimeout:  time.Duration(conf.RequestTimeout) * time.Second,
		WriteTimeout: time.Duration(conf.RequestTimeout) * time.Second,
		TLSConfig: &tls.Config{
			NextProtos: []string{"h3", "h2", "http/1.1"},
		},
	}
}

// startServerInstance starts the HTTP server instance.
func startServerInstance(server *http.Server, conf config.SiteConfig, l *logger.Logger) {
	go func() {
		if conf.SSL.Enabled {
			// SSL/TLS configuration
			if _, err := os.Stat(conf.SSL.Certificate); os.IsNotExist(err) {
				l.Errorf("SSL certificate not found for %s: %v", conf.Domain, err)
				return
			}
			if _, err := os.Stat(conf.SSL.Key); os.IsNotExist(err) {
				l.Errorf("SSL key not found for %s: %v", conf.Domain, err)
				return
			}

			l.Infof("Serving %s on HTTPS port %d with HTTP/2 and HTTP/3 support", conf.Domain, conf.Port)

			// HTTP/1.1 and HTTP/2 server are also started to keep compatibility
			// with clients that do not support HTTP/3
			go func() {
				if err := server.ListenAndServeTLS(conf.SSL.Certificate, conf.SSL.Key); err != nil && err != http.ErrServerClosed {
					l.Errorf("HTTP/1.1 and HTTP/2 server error for %s: %v", conf.Domain, err)
				}
			}()

			quicAddr := fmt.Sprintf(":%d", conf.Port)
			err := http3.ListenAndServeQUIC(quicAddr, conf.SSL.Certificate, conf.SSL.Key, server.Handler)
			if err != nil && err != http.ErrServerClosed {
				l.Errorf("HTTP/3 server error for %s: %v", conf.Domain, err)
			}
		} else {
			l.Infof("Serving on HTTP port %d", conf.Port)
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				l.Errorf("Server error on port %d: %v", conf.Port, err)
			}
		}
	}()
}
