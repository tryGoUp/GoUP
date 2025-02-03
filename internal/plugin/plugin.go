package plugin

import (
	"sync"

	"github.com/mirkobrombin/goup/internal/config"
	"github.com/mirkobrombin/goup/internal/server/middleware"
	log "github.com/sirupsen/logrus"
)

// Plugin defines the interface for GoUP plugins.
type Plugin interface {
	Name() string
	Init(mwManager *middleware.MiddlewareManager) error
	InitForSite(mwManager *middleware.MiddlewareManager, logger *log.Logger, conf config.SiteConfig) error
}

// PluginManager manages loading and initialization of plugins.
type PluginManager struct {
	plugins []Plugin
	mu      sync.Mutex
}

// DefaultPluginManager is the default instance used by the application.
var DefaultPluginManager *PluginManager

// NewPluginManager creates a new PluginManager instance.
func NewPluginManager() *PluginManager {
	return &PluginManager{
		plugins: []Plugin{},
	}
}

// SetDefaultPluginManager sets the default PluginManager instance.
func SetDefaultPluginManager(pm *PluginManager) {
	DefaultPluginManager = pm
}

// GetPluginManagerInstance returns the default PluginManager instance.
// If it is not set, a new one is created.
func GetPluginManagerInstance() *PluginManager {
	if DefaultPluginManager == nil {
		DefaultPluginManager = NewPluginManager()
	}
	return DefaultPluginManager
}

// Register registers a new plugin.
func (pm *PluginManager) Register(plugin Plugin) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.plugins = append(pm.plugins, plugin)
}

// InitPlugins initializes all registered plugins.
func (pm *PluginManager) InitPlugins(mwManager *middleware.MiddlewareManager) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for _, plugin := range pm.plugins {
		if err := plugin.Init(mwManager); err != nil {
			return err
		}
	}
	return nil
}

// InitPluginsForSite initializes all registered plugins for a specific site.
func (pm *PluginManager) InitPluginsForSite(mwManager *middleware.MiddlewareManager, baseLogger *log.Logger, conf config.SiteConfig) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for _, plugin := range pm.plugins {
		pluginLogger := baseLogger.WithFields(log.Fields{
			"plugin": plugin.Name(),
			"domain": conf.Domain,
		})

		if err := plugin.InitForSite(mwManager, pluginLogger.Logger, conf); err != nil {
			return err
		}
	}
	return nil
}

// GetRegisteredPlugins returns the names of all registered plugins.
func (pm *PluginManager) GetRegisteredPlugins() []string {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	names := make([]string, len(pm.plugins))
	for i, plugin := range pm.plugins {
		names[i] = plugin.Name()
	}
	return names
}
