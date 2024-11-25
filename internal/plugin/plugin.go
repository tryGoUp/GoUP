package plugin

import (
	"sync"

	"github.com/mirkobrombin/goup/internal/server/middleware"
)

// Plugin defines the interface for GoUP plugins.
type Plugin interface {
	Name() string
	Init(mwManager *middleware.MiddlewareManager) error
}

// PluginManager manages loading and initialization of plugins.
type PluginManager struct {
	plugins []Plugin
	mu      sync.Mutex
}

// Singleton PluginManagerInstance.
var pluginManagerInstance *PluginManager
var once sync.Once

// GetPluginManagerInstance returns the singleton instance of PluginManager.
func GetPluginManagerInstance() *PluginManager {
	once.Do(func() {
		pluginManagerInstance = &PluginManager{
			plugins: []Plugin{},
		}
	})
	return pluginManagerInstance
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

// GetRegisteredPlugins returns the registered plugins.
func (pm *PluginManager) GetRegisteredPlugins() []string {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	names := make([]string, len(pm.plugins))
	for i, plugin := range pm.plugins {
		names[i] = plugin.Name()
	}
	return names
}
