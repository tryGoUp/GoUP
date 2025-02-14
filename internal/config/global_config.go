package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// GlobalConfig contains the global settings for GoUP.
type GlobalConfig struct {
	EnableAPI      bool     `json:"enable_api"`
	APIPort        int      `json:"api_port"`
	DashboardPort  int      `json:"dashboard_port"`
	EnabledPlugins []string `json:"enabled_plugins"` // empty means all enabled
}

// GlobalConf is the global configuration in memory.
var GlobalConf *GlobalConfig
var globalConfName = "conf.global.json"

// LoadGlobalConfig loads the global configuration file.
func LoadGlobalConfig() error {
	configDir := GetConfigDir()
	configFile := filepath.Join(configDir, globalConfName)
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		GlobalConf = &GlobalConfig{
			EnableAPI:      true,
			APIPort:        6007,
			DashboardPort:  6008,
			EnabledPlugins: []string{},
		}
		return nil
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}
	var conf GlobalConfig
	if err := json.Unmarshal(data, &conf); err != nil {
		return err
	}
	GlobalConf = &conf
	return nil
}

// SaveGlobalConfig saves the global configuration file.
func SaveGlobalConfig() error {
	configDir := GetConfigDir()
	configFile := filepath.Join(configDir, globalConfName)
	data, err := json.MarshalIndent(GlobalConf, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(configFile, data, 0644)
}
