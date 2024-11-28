package plugins

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/mirkobrombin/goup/internal/config"
	"github.com/mirkobrombin/goup/internal/server/middleware"
	log "github.com/sirupsen/logrus"
	"github.com/yookoala/gofast"
)

// PHPPlugin handles the execution of PHP scripts via PHP-FPM.
type PHPPlugin struct{}

// Name returns the name of the plugin.
func (p *PHPPlugin) Name() string {
	return "PHPPlugin"
}

// Init registers any global middleware (none for PHPPlugin).
func (p *PHPPlugin) Init(mwManager *middleware.MiddlewareManager) error {
	return nil
}

// InitForSite initializes the plugin for a specific site.
func (p *PHPPlugin) InitForSite(mwManager *middleware.MiddlewareManager, logger *log.Logger, conf config.SiteConfig) error {
	mwManager.Use(p.phpMiddleware(logger, conf))
	return nil
}

// PHPPluginConfig represents the configuration for the PHPPlugin.
type PHPPluginConfig struct {
	Enable  bool   `json:"enable"`
	FPMAddr string `json:"fpm_addr"`
}

// phpMiddleware is the middleware that handles PHP requests.
func (p *PHPPlugin) phpMiddleware(logger *log.Logger, conf config.SiteConfig) middleware.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// Retrieve site-specific plugin configuration.
			pluginConfigRaw, ok := conf.PluginConfigs[p.Name()]
			if !ok {
				logger.Warnf("Plugin config not found for host: %s", r.Host)
				next.ServeHTTP(w, r)
				return
			}

			// FIXME: find a better way to map the configuration, currently
			// it is like this due to fpm_addr not being unmarshalled correctly.
			pluginConfig := PHPPluginConfig{}
			if rawMap, ok := pluginConfigRaw.(map[string]interface{}); ok {
				if enable, ok := rawMap["enable"].(bool); ok {
					pluginConfig.Enable = enable
				}
				if fpmAddr, ok := rawMap["fpm_addr"].(string); ok {
					pluginConfig.FPMAddr = fpmAddr
				}
			}

			if !pluginConfig.Enable {
				logger.Infof("PHP Plugin is disabled for host: %s", r.Host)
				next.ServeHTTP(w, r)
				return
			}

			// We only handle PHP requests here.
			if strings.HasSuffix(r.URL.Path, ".php") {
				logger.Infof("Handling PHP request: %s", r.URL.Path)

				phpFPMAddr := pluginConfig.FPMAddr
				if phpFPMAddr == "" {
					phpFPMAddr = "127.0.0.1:9000"
				}

				var connFactory gofast.ConnFactory
				if strings.HasPrefix(phpFPMAddr, "/") {
					connFactory = gofast.SimpleConnFactory("unix", phpFPMAddr)
				} else {
					connFactory = gofast.SimpleConnFactory("tcp", phpFPMAddr)
				}

				clientFactory := gofast.SimpleClientFactory(connFactory)

				// Build the full path to the script
				root := conf.RootDirectory
				scriptFilename := filepath.Join(root, r.URL.Path)
				if _, err := os.Stat(scriptFilename); os.IsNotExist(err) {
					http.NotFound(w, r)
					return
				}

				// Create a new FastCGI handler
				fcgiHandler := gofast.NewHandler(
					func(client gofast.Client, req *gofast.Request) (*gofast.ResponsePipe, error) {
						req.Params["SCRIPT_FILENAME"] = scriptFilename
						req.Params["DOCUMENT_ROOT"] = root
						req.Params["REQUEST_METHOD"] = r.Method
						req.Params["SERVER_PROTOCOL"] = r.Proto
						req.Params["REQUEST_URI"] = r.URL.RequestURI()
						req.Params["QUERY_STRING"] = r.URL.RawQuery
						req.Params["REMOTE_ADDR"] = r.RemoteAddr

						return gofast.BasicSession(client, req)
					},
					clientFactory,
				)

				// Serve the request
				fcgiHandler.ServeHTTP(w, r)
			} else {
				next.ServeHTTP(w, r)
			}
		})
	}
}
