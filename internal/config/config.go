package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
)

// SSLConfig represents the SSL configuration for a site.
type SSLConfig struct {
	Enabled     bool   `json:"enabled"`
	Certificate string `json:"certificate"`
	Key         string `json:"key"`
}

// SiteConfig contains the configuration for a single site.
type SiteConfig struct {
	Domain         string            `json:"domain"`
	Port           int               `json:"port"`
	RootDirectory  string            `json:"root_directory"`
	CustomHeaders  map[string]string `json:"custom_headers"`
	ProxyPass      string            `json:"proxy_pass"`
	SSL            SSLConfig         `json:"ssl"`
	RequestTimeout int               `json:"request_timeout"` // in seconds
}

// GetConfigDir returns the directory where configuration files are stored.
func GetConfigDir() string {
	var configDir string
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		configDir = filepath.Join(xdgConfig, "goup")
	} else if runtime.GOOS == "windows" {
		configDir = filepath.Join(os.Getenv("APPDATA"), "goup")
	} else {
		configDir = filepath.Join(os.Getenv("HOME"), ".config", "goup")
	}
	return configDir
}

// GetLogDir returns the directory where log files are stored.
func GetLogDir() string {
	var logDir string
	if xdgDataHome := os.Getenv("XDG_DATA_HOME"); xdgDataHome != "" {
		logDir = filepath.Join(xdgDataHome, "goup", "logs")
	} else if runtime.GOOS == "windows" {
		logDir = filepath.Join(os.Getenv("APPDATA"), "goup", "logs")
	} else {
		logDir = filepath.Join(os.Getenv("HOME"), ".local", "share", "goup", "logs")
	}
	return logDir
}

// LoadConfig loads a configuration from a file.
func LoadConfig(filePath string) (SiteConfig, error) {
	var conf SiteConfig
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return conf, err
	}
	if err := json.Unmarshal(data, &conf); err != nil {
		return conf, err
	}
	return conf, nil
}

// LoadAllConfigs loads all configurations from the configuration directory.
func LoadAllConfigs() ([]SiteConfig, error) {
	configDir := GetConfigDir()
	files, err := ioutil.ReadDir(configDir)
	if err != nil {
		return nil, err
	}

	var configs []SiteConfig
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			conf, err := LoadConfig(filepath.Join(configDir, file.Name()))
			if err != nil {
				fmt.Printf("Error loading config %s: %v\n", file.Name(), err)
				continue
			}
			configs = append(configs, conf)
		}
	}
	return configs, nil
}

// Save saves the configuration to a file.
func (conf *SiteConfig) Save(filePath string) error {
	data, err := json.MarshalIndent(conf, "", "    ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filePath, data, 0644)
}
