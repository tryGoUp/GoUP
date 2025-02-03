package plugins

import (
	"net/http"
	"strings"

	"github.com/mirkobrombin/goup/internal/config"
	log "github.com/sirupsen/logrus"
)

// CustomHeaderPlugin adds a custom header to all HTTP responses.
type CustomHeaderPlugin struct {
	// siteConfigs stores the configurations keyed by domain.
	siteConfigs map[string]config.SiteConfig
	// logger is optional and can be shared among sites.
	logger *log.Logger
}

// Name returns the name of the plugin.
func (p *CustomHeaderPlugin) Name() string {
	return "CustomHeaderPlugin"
}

// OnInit is called once during the global plugin initialization.
func (p *CustomHeaderPlugin) OnInit() error {
	p.siteConfigs = make(map[string]config.SiteConfig)
	return nil
}

// OnInitForSite initializes the plugin for a specific site.
func (p *CustomHeaderPlugin) OnInitForSite(conf config.SiteConfig, logger *log.Logger) error {
	// Store the config, keyed by domain.
	p.siteConfigs[conf.Domain] = conf
	if p.logger == nil {
		p.logger = logger
	}
	return nil
}

// BeforeRequest is invoked before serving each request.
func (p *CustomHeaderPlugin) BeforeRequest(r *http.Request) {
	// No operation here; logic is in HandleRequest.
}

// HandleRequest can fully handle the request, returning true if it does so.
func (p *CustomHeaderPlugin) HandleRequest(w http.ResponseWriter, r *http.Request) bool {
	// Identify the domain and strip any port.
	host := r.Host
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}

	// Check if there's a site config for this domain.
	if conf, ok := p.siteConfigs[host]; ok {
		for key, value := range conf.CustomHeaders {
			w.Header().Set(key, value)
			if p.logger != nil {
				p.logger.Infof("Added custom header %s: %s", key, value)
			}
		}
	}
	// Returning false tells GoUP to continue with other handlers.
	return false
}

// AfterRequest is invoked after the request has been served or handled.
func (p *CustomHeaderPlugin) AfterRequest(w http.ResponseWriter, r *http.Request) {
	// No operation here.
}

// OnExit is called when the server is shutting down.
func (p *CustomHeaderPlugin) OnExit() error {
	return nil
}
