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
	"github.com/mirkobrombin/goup/internal/tools"
	"github.com/mirkobrombin/goup/internal/tui"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

var (
	loggers = make(map[string]*logger.Logger)
	tuiMode bool
)

// StartServers starts the servers based on the provided configurations.
func StartServers(configs []config.SiteConfig, enableTUI bool, enableBench bool) {
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
		fields := logger.Fields{"domain": identifier}
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

	// Enable benchmark middleware if requested
	if enableBench {
		mwManager.Use(middleware.BenchmarkMiddleware())
	}

	// Initialize the plugins globally
	pluginManager := plugin.GetPluginManagerInstance()
	if err := pluginManager.InitPlugins(); err != nil {
		fmt.Printf("Error initializing plugins: %v\n", err)
		return
	}

	var wg sync.WaitGroup

	for port, confs := range portConfigs {
		wg.Add(1)
		go func(port int, confs []config.SiteConfig) {
			defer wg.Done()
			if len(confs) == 1 {
				conf := confs[0]
				startSingleServer(conf, mwManager, pluginManager)
			} else {
				startVirtualHostServer(port, confs, mwManager, pluginManager)
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

func anyHasSSL(confs []config.SiteConfig) bool {
	for _, c := range confs {
		if c.SSL.Enabled {
			return true
		}
	}
	return false
}

// startSingleServer starts a server for a single site configuration.
func startSingleServer(conf config.SiteConfig, mwManager *middleware.MiddlewareManager, pm *plugin.PluginManager) {
	identifier := conf.Domain
	lg := loggers[identifier]

	// We do not want to start a server if the root directory does not exist
	// let's fail fast instead.
	if conf.ProxyPass == "" {
		// Here we allow empty paths as RootDirectory, this is useful for
		// proxying requests to other servers by default, like Flask apps.
		if conf.RootDirectory != "" {
			if _, err := os.Stat(conf.RootDirectory); os.IsNotExist(err) {
				lg.Errorf("Root directory does not exist for %s: %v", conf.Domain, err)
				return
			}
		}
	}

	// Initialize plugins for this site
	if err := pm.InitPluginsForSite(conf, lg); err != nil {
		lg.Errorf("Error initializing plugins for site %s: %v", conf.Domain, err)
		return
	}

	// Add plugin middleware
	mwManagerCopy := mwManager.Copy()
	mwManagerCopy.Use(plugin.PluginMiddleware(pm))

	handler, err := createHandler(conf, lg, identifier, mwManagerCopy)
	if err != nil {
		lg.Errorf("Error creating handler for %s: %v", conf.Domain, err)
		return
	}

	// If SSL is enabled, keep the original net/http + quic-go approach, since
	// fasthttp does not support QUIC (yet?).
	if conf.SSL.Enabled {
		server := createHTTPServer(conf, handler)
		startServerInstance(server, conf, lg)
	} else {
		startFasthttpServer(conf, handler, lg)
	}
}

// startVirtualHostServer starts a server that handles multiple domains on the same port.
func startVirtualHostServer(port int, configs []config.SiteConfig, mwManager *middleware.MiddlewareManager, pm *plugin.PluginManager) {
	identifier := fmt.Sprintf("port_%d", port)
	lg := loggers[identifier]

	// If any of the sites has SSL enabled, we need to use the net/http server.
	if anyHasSSL(configs) {
		radixTree := radix.New()

		for _, conf := range configs {
			if conf.ProxyPass == "" && conf.RootDirectory != "" {
				if _, err := os.Stat(conf.RootDirectory); os.IsNotExist(err) {
					lg.Errorf("Root directory does not exist for %s: %v", conf.Domain, err)
				}
			}

			if err := pm.InitPluginsForSite(conf, lg); err != nil {
				lg.Errorf("Error initializing plugins for site %s: %v", conf.Domain, err)
				continue
			}

			mwManagerCopy := mwManager.Copy()
			mwManagerCopy.Use(plugin.PluginMiddleware(pm))

			handler, err := createHandler(conf, lg, identifier, mwManagerCopy)
			if err != nil {
				lg.Errorf("Error creating handler for %s: %v", conf.Domain, err)
				continue
			}

			radixTree.Insert(conf.Domain, handler)
		}

		serverConf := config.SiteConfig{Port: port}

		mainHandler := func(w_ http.ResponseWriter, r_ *http.Request) {
			host := r_.Host
			if colonIndex := strings.Index(host, ":"); colonIndex != -1 {
				host = host[:colonIndex]
			}
			if h, found := radixTree.Get(host); found {
				h.(http.Handler).ServeHTTP(w_, r_)
			} else {
				http.NotFound(w_, r_)
			}
		}

		server := createHTTPServer(serverConf, http.HandlerFunc(mainHandler))
		startServerInstance(server, serverConf, lg)
	} else {
		// fasthttp for all other cases.
		radixTree := radix.New()

		for _, conf := range configs {
			if conf.ProxyPass == "" && conf.RootDirectory != "" {
				if _, err := os.Stat(conf.RootDirectory); os.IsNotExist(err) {
					lg.Errorf("Root directory does not exist for %s: %v", conf.Domain, err)
				}
			}

			if err := pm.InitPluginsForSite(conf, lg); err != nil {
				lg.Errorf("Error initializing plugins for site %s: %v", conf.Domain, err)
				continue
			}

			mwManagerCopy := mwManager.Copy()
			mwManagerCopy.Use(plugin.PluginMiddleware(pm))

			nethttpHandler, err := createHandler(conf, lg, identifier, mwManagerCopy)
			if err != nil {
				lg.Errorf("Error creating handler for %s: %v", conf.Domain, err)
				continue
			}

			fasthttpHandler := fasthttpadaptor.NewFastHTTPHandler(nethttpHandler)
			radixTree.Insert(conf.Domain, fasthttpHandler)
		}

		fasthttpMainHandler := func(ctx *fasthttp.RequestCtx) {
			host := string(ctx.Host())
			if colonIndex := strings.Index(host, ":"); colonIndex != -1 {
				host = host[:colonIndex]
			}

			if h, found := radixTree.Get(host); found {
				h.(fasthttp.RequestHandler)(ctx)
			} else {
				ctx.SetStatusCode(fasthttp.StatusNotFound)
			}
		}

		serverConf := config.SiteConfig{Port: port}

		server := &fasthttp.Server{
			Handler:      fasthttpMainHandler,
			ReadTimeout:  tools.TimeDurationOrDefault(serverConf.RequestTimeout),
			WriteTimeout: tools.TimeDurationOrDefault(serverConf.RequestTimeout),
		}

		lg.Infof("Serving on HTTP port %d with fasthttp", port)
		err := server.ListenAndServe(fmt.Sprintf(":%d", port))
		if err != nil {
			lg.Errorf("Fasthttp server error on port %d: %v", port, err)
		}
	}
}

// startFasthttpServer starts a fasthttp server for the given site configuration.
func startFasthttpServer(conf config.SiteConfig, nethttpHandler http.Handler, lg *logger.Logger) {
	fasthttpHandler := fasthttpadaptor.NewFastHTTPHandler(nethttpHandler)

	server := &fasthttp.Server{
		Handler:      fasthttpHandler,
		ReadTimeout:  tools.TimeDurationOrDefault(conf.RequestTimeout),
		WriteTimeout: tools.TimeDurationOrDefault(conf.RequestTimeout),
	}

	lg.Infof("Serving on HTTP port %d with fasthttp", conf.Port)
	err := server.ListenAndServe(fmt.Sprintf(":%d", conf.Port))
	if err != nil {
		lg.Errorf("Fasthttp server error on port %d: %v", conf.Port, err)
	}
}
