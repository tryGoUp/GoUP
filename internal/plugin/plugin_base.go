package plugin

import (
	"github.com/mirkobrombin/goup/internal/config"
	"github.com/mirkobrombin/goup/internal/logger"
	log "github.com/sirupsen/logrus"
)

// BasePlugin stores two loggers and a domain name.
// - DomainLogger: logs to console + domain file
// - PluginLogger: logs only to the plugin-specific file
type BasePlugin struct {
	DomainLogger *log.Logger
	PluginLogger *log.Logger
	Domain       string
}

// SetupLoggers is typically called in OnInitForSite, so each plugin can
// use DomainLogger and PluginLogger without rewriting the same logic.
func (bp *BasePlugin) SetupLoggers(conf config.SiteConfig, pluginName string, domainLogger *log.Logger) error {
	pluginLogger, err := logger.NewPluginLogger(conf.Domain, pluginName)
	if err != nil {
		domainLogger.Errorf("Failed to create plugin logger for domain %s: %v", conf.Domain, err)
		return err
	}

	bp.DomainLogger = domainLogger
	bp.PluginLogger = pluginLogger
	bp.Domain = conf.Domain

	return nil
}
