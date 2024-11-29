package server

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/armon/go-radix"
	"github.com/mirkobrombin/goup/internal/config"
	"github.com/mirkobrombin/goup/internal/logger"
	"github.com/mirkobrombin/goup/internal/plugin"
	"github.com/mirkobrombin/goup/internal/server/middleware"
	"github.com/mirkobrombin/goup/internal/tui"
	log "github.com/sirupsen/logrus"
)

var (
	loggers = make(map[string]*log.Logger)
	tuiMode bool
)

// StartServers starts the servers based on the provided configurations.
func StartServers(configs []config.SiteConfig, enableTUI bool) {
	tuiMode = enableTUI

	// FIXME: move all TUI related code out of this package, I do not feel
	// comfortable having it here, leads to confusion.
	if tuiMode {
		tui.InitTUI()
	}

	// Groupping configurations by port to minimize the number of servers
	// NOTE: configurations with the same port are treated as virtual hosts
	// so they will be served by the same server instance.
	portConfigs := make(map[int][]config.SiteConfig)
	for _, conf := range configs {
		portConfigs[conf.Port] = append(portConfigs[conf.Port], conf)
	}

	// Setting up loggers and TUI views before starting servers so that
	// they are ready to host the messages.
	for port, confs := range portConfigs {
		var identifier string
		if len(confs) == 1 {
			identifier = confs[0].Domain
		} else {
			identifier = fmt.Sprintf("port_%d", port)
		}

		// Set up logger
		fields := log.Fields{
			"domain": identifier,
		}
		lg, err := logger.NewLogger(identifier, fields)
		if err != nil {
			fmt.Printf("Error setting up logger for %s: %v\n", identifier, err)
			continue
		}
		loggers[identifier] = lg

		// Set up TUI view
		if tuiMode {
			tui.SetupView(identifier)
		}
	}

	// Initialize the global middleware manager
	mwManager := middleware.NewMiddlewareManager()

	// Initialize the plugins with the global middleware manager
	pluginManager := plugin.GetPluginManagerInstance()
	if err := pluginManager.InitPlugins(mwManager); err != nil {
		fmt.Printf("Error initializing plugins: %v\n", err)
		return
	}

	var wg sync.WaitGroup

	for port, confs := range portConfigs {
		wg.Add(1)
		go func(port int, confs []config.SiteConfig) {
			defer wg.Done()
			if len(confs) == 1 {
				// Single domain on this port, start dedicated server
				conf := confs[0]
				startSingleServer(conf, mwManager)
			} else {
				// Multiple domains on this port, start server with
				// virtual host support.
				startVirtualHostServer(port, confs, mwManager)
			}
		}(port, confs)
	}

	// Start TUI if enabled
	if tuiMode {
		tui.Run()
	} else {
		// Let's keep alive the main goroutine alive
		wg.Wait()
	}
}

// startSingleServer starts a server for a single site configuration.
func startSingleServer(conf config.SiteConfig, mwManager *middleware.MiddlewareManager) {
	identifier := conf.Domain
	logger := loggers[identifier]

	// We do not want to start a server if the root directory does not exist
	// let's fail fast instead.
	if conf.ProxyPass == "" {
		if _, err := os.Stat(conf.RootDirectory); os.IsNotExist(err) {
			logger.Errorf("Root directory does not exist for %s: %v", conf.Domain, err)
			return
		}
	}

	handler, err := createHandler(conf, logger, identifier, mwManager)
	if err != nil {
		logger.Errorf("Error creating handler for %s: %v", conf.Domain, err)
		return
	}

	server := createHTTPServer(conf, handler)
	startServerInstance(server, conf, logger)
}

// startVirtualHostServer starts a server that handles multiple domains on the same port.
func startVirtualHostServer(port int, configs []config.SiteConfig, mwManager *middleware.MiddlewareManager) {
	identifier := fmt.Sprintf("port_%d", port)
	logger := loggers[identifier]

	radixTree := radix.New()

	for _, conf := range configs {
		// We do not want to start a server if the root directory does not exist
		// let's fail fast instead.
		if conf.ProxyPass == "" {
			if _, err := os.Stat(conf.RootDirectory); os.IsNotExist(err) {
				logger.Errorf("Root directory does not exist for %s: %v", conf.Domain, err)
			}
		}

		handler, err := createHandler(conf, logger, identifier, mwManager)
		if err != nil {
			logger.Errorf("Error creating handler for %s: %v", conf.Domain, err)
			continue
		}

		radixTree.Insert(conf.Domain, handler)
	}

	// Main handler that routes requests based on the Host header
	mainHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := r.Host
		if colonIndex := strings.Index(host, ":"); colonIndex != -1 {
			host = host[:colonIndex]
		}

		if handler, found := radixTree.Get(host); found {
			handler.(http.Handler).ServeHTTP(w, r)
		} else {
			http.NotFound(w, r)
		}
	})

	serverConf := config.SiteConfig{
		Port: port,
	}
	server := createHTTPServer(serverConf, mainHandler)
	startServerInstance(server, serverConf, logger)
}
