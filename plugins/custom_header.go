package plugins

import (
	"net/http"
	"strings"

	"github.com/mirkobrombin/goup/internal/config"
	"github.com/mirkobrombin/goup/internal/logger"
	"github.com/mirkobrombin/goup/internal/plugin"
)

// CustomHeaderPlugin adds a custom header to all HTTP responses.
type CustomHeaderPlugin struct {
	plugin.BasePlugin

	siteConfigs map[string]config.SiteConfig
}

func (p *CustomHeaderPlugin) Name() string {
	return "CustomHeaderPlugin"
}

func (p *CustomHeaderPlugin) OnInit() error {
	p.siteConfigs = make(map[string]config.SiteConfig)
	return nil
}

func (p *CustomHeaderPlugin) OnInitForSite(conf config.SiteConfig, domainLogger *logger.Logger) error {
	if err := p.SetupLoggers(conf, p.Name(), domainLogger); err != nil {
		return err
	}
	p.siteConfigs[conf.Domain] = conf
	return nil
}

func (p *CustomHeaderPlugin) BeforeRequest(r *http.Request) {}

func (p *CustomHeaderPlugin) HandleRequest(w http.ResponseWriter, r *http.Request) bool {
	// Identify the domain and strip any port.
	host := r.Host
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}

	// Check if there's a site config for this domain.
	conf, ok := p.siteConfigs[host]
	if !ok {
		return false
	}

	for key, value := range conf.CustomHeaders {
		w.Header().Set(key, value)
		p.DomainLogger.Infof("[CustomHeaderPlugin] Added header=%s value=%s (domain=%s)", key, value, host)
	}
	return false
}

func (p *CustomHeaderPlugin) AfterRequest(w http.ResponseWriter, r *http.Request) {}
func (p *CustomHeaderPlugin) OnExit() error                                       { return nil }
