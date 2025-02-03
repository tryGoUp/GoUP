package plugins

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/mirkobrombin/goup/internal/config"
	log "github.com/sirupsen/logrus"
	"github.com/yookoala/gofast"
)

// PHPPlugin handles the execution of PHP scripts via PHP-FPM.
type PHPPlugin struct {
	logger      *log.Logger
	siteConfigs map[string]PHPPluginConfig
}

// PHPPluginConfig represents the configuration for the PHPPlugin.
type PHPPluginConfig struct {
	Enable  bool   `json:"enable"`
	FPMAddr string `json:"fpm_addr"`
}

// Name returns the name of the plugin.
func (p *PHPPlugin) Name() string {
	return "PHPPlugin"
}

// OnInit registers any global plugin logic (none in this case).
func (p *PHPPlugin) OnInit() error {
	p.siteConfigs = make(map[string]PHPPluginConfig)
	return nil
}

// OnInitForSite initializes the plugin for a specific site.
func (p *PHPPlugin) OnInitForSite(conf config.SiteConfig, logger *log.Logger) error {
	if p.logger == nil {
		p.logger = logger
	}

	// Retrieve plugin config from conf.PluginConfigs
	pluginConfigRaw, ok := conf.PluginConfigs[p.Name()]
	if !ok {
		// No config for PHP, store default disabled config.
		p.siteConfigs[conf.Domain] = PHPPluginConfig{}
		return nil
	}

	cfg := PHPPluginConfig{}
	if rawMap, ok := pluginConfigRaw.(map[string]interface{}); ok {
		if enable, ok := rawMap["enable"].(bool); ok {
			cfg.Enable = enable
		}
		if fpmAddr, ok := rawMap["fpm_addr"].(string); ok {
			cfg.FPMAddr = fpmAddr
		}
	}
	p.siteConfigs[conf.Domain] = cfg

	return nil
}

// BeforeRequest is invoked before serving each request (unused here).
func (p *PHPPlugin) BeforeRequest(r *http.Request) {}

// HandleRequest can fully handle the request, returning true if it does so.
func (p *PHPPlugin) HandleRequest(w http.ResponseWriter, r *http.Request) bool {
	host := r.Host
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}

	cfg, ok := p.siteConfigs[host]
	if !ok || !cfg.Enable {
		return false
	}

	// We only handle .php files.
	if !strings.HasSuffix(r.URL.Path, ".php") {
		return false
	}

	p.logger.Infof("Handling PHP request: %s", r.URL.Path)

	// If the user hasn't specified a FPM address, use default.
	phpFPMAddr := cfg.FPMAddr
	if phpFPMAddr == "" {
		phpFPMAddr = "127.0.0.1:9000"
	}

	rootDir := p.getRootDirectory(r) // We'll retrieve it from somewhere or do a fallback.

	scriptFilename := filepath.Join(rootDir, r.URL.Path)
	if _, err := os.Stat(scriptFilename); os.IsNotExist(err) {
		http.NotFound(w, r)
		return true
	}

	var connFactory gofast.ConnFactory
	if strings.HasPrefix(phpFPMAddr, "/") {
		connFactory = gofast.SimpleConnFactory("unix", phpFPMAddr)
	} else {
		connFactory = gofast.SimpleConnFactory("tcp", phpFPMAddr)
	}

	clientFactory := gofast.SimpleClientFactory(connFactory)

	fcgiHandler := gofast.NewHandler(
		func(client gofast.Client, req *gofast.Request) (*gofast.ResponsePipe, error) {
			req.Params["SCRIPT_FILENAME"] = scriptFilename
			req.Params["DOCUMENT_ROOT"] = rootDir
			req.Params["REQUEST_METHOD"] = r.Method
			req.Params["SERVER_PROTOCOL"] = r.Proto
			req.Params["REQUEST_URI"] = r.URL.RequestURI()
			req.Params["QUERY_STRING"] = r.URL.RawQuery
			req.Params["REMOTE_ADDR"] = r.RemoteAddr
			return gofast.BasicSession(client, req)
		},
		clientFactory,
	)

	fcgiHandler.ServeHTTP(w, r)
	return true
}

// AfterRequest is invoked after the request has been served or handled.
func (p *PHPPlugin) AfterRequest(w http.ResponseWriter, r *http.Request) {}

// OnExit is called when the server is shutting down.
func (p *PHPPlugin) OnExit() error {
	return nil
}

// getRootDirectory tries to derive the site root from the request.
// If there's a site-specific approach, do it here; otherwise, fallback.
func (p *PHPPlugin) getRootDirectory(r *http.Request) string {
	// TODO: If needed, store site root in a domain->root map, or parse from
	//       the request. For now, fallback to a default or just return ".".
	return "."
}
