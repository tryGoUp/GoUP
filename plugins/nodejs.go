package plugins

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/mirkobrombin/goup/internal/config"
	"github.com/mirkobrombin/goup/internal/server/middleware"
	log "github.com/sirupsen/logrus"
)

// NodeJSPlugin handles the execution of a Node.js application.
type NodeJSPlugin struct {
	mu      sync.Mutex
	process *os.Process
}

// Name returns the name of the plugin.
func (p *NodeJSPlugin) Name() string {
	return "NodeJSPlugin"
}

// Init registers any global middleware (none for NodeJSPlugin).
func (p *NodeJSPlugin) Init(mwManager *middleware.MiddlewareManager) error {
	return nil
}

// InitForSite initializes the plugin for a specific site.
func (p *NodeJSPlugin) InitForSite(mwManager *middleware.MiddlewareManager, logger *log.Logger, conf config.SiteConfig) error {
	mwManager.Use(p.nodeMiddleware(logger, conf))
	return nil
}

// NodeJSPluginConfig represents the configuration for the NodeJSPlugin.
type NodeJSPluginConfig struct {
	Enable         bool     `json:"enable"`
	Port           string   `json:"port"`
	RootDir        string   `json:"root_dir"`
	Entry          string   `json:"entry"`
	InstallDeps    bool     `json:"install_deps"`
	NodePath       string   `json:"node_path"`
	PackageManager string   `json:"package_manager"`
	ProxyPaths     []string `json:"proxy_paths"`
}

// nodeMiddleware intercepts requests and forwards API calls to Node.js.
func (p *NodeJSPlugin) nodeMiddleware(logger *log.Logger, conf config.SiteConfig) middleware.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Retrieve site-specific plugin configuration.
			pluginConfigRaw, ok := conf.PluginConfigs[p.Name()]
			if !ok {
				logger.Warnf("Plugin config not found for host: %s", r.Host)
				next.ServeHTTP(w, r)
				return
			}

			// Parse configuration.
			pluginConfig := NodeJSPluginConfig{}
			if rawMap, ok := pluginConfigRaw.(map[string]interface{}); ok {
				if enable, ok := rawMap["enable"].(bool); ok {
					pluginConfig.Enable = enable
				}
				if port, ok := rawMap["port"].(string); ok {
					pluginConfig.Port = port
				}
				if rootDir, ok := rawMap["root_dir"].(string); ok {
					pluginConfig.RootDir = rootDir
				}
				if entry, ok := rawMap["entry"].(string); ok {
					pluginConfig.Entry = entry
				}
				if installDeps, ok := rawMap["install_deps"].(bool); ok {
					pluginConfig.InstallDeps = installDeps
				}
				if nodePath, ok := rawMap["node_path"].(string); ok {
					pluginConfig.NodePath = nodePath
				}
				if packageManager, ok := rawMap["package_manager"].(string); ok {
					pluginConfig.PackageManager = packageManager
				}
				if proxyPaths, ok := rawMap["proxy_paths"].([]interface{}); ok {
					for _, path := range proxyPaths {
						if pathStr, ok := path.(string); ok {
							pluginConfig.ProxyPaths = append(pluginConfig.ProxyPaths, pathStr)
						}
					}
				}
			}

			if !pluginConfig.Enable {
				logger.Infof("NodeJS Plugin is disabled for host: %s", r.Host)
				next.ServeHTTP(w, r)
				return
			}

			// Start Node.js if it is not already running.
			p.ensureNodeServerRunning(pluginConfig, logger)

			// Check if the request should be forwarded to Node.js.
			for _, proxyPath := range pluginConfig.ProxyPaths {
				if strings.HasPrefix(r.URL.Path, proxyPath) {
					p.proxyToNode(w, r, pluginConfig, logger)
					return
				}
			}

			// If the request does not match a Node.js route, let GoUp handle static files.
			next.ServeHTTP(w, r)
		})
	}
}

// ensureNodeServerRunning starts Node.js if it is not already running.
func (p *NodeJSPlugin) ensureNodeServerRunning(config NodeJSPluginConfig, logger *log.Logger) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// If the process is already running, do nothing.
	if p.process != nil {
		logger.Infof("Node.js server is already running (PID: %d)", p.process.Pid)
		return
	}

	logger.Infof("Starting Node.js server...")

	// Install dependencies if required.
	if config.InstallDeps {
		p.installDependencies(config, logger)
	}

	// Start the Node.js server.
	entryPath := filepath.Join(config.RootDir, config.Entry)

	nodePath := config.NodePath
	if nodePath == "" {
		nodePath = "node"
	}

	cmd := exec.Command(nodePath, entryPath)
	cmd.Dir = config.RootDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		logger.Errorf("Failed to start Node.js server: %v", err)
		return
	}

	p.process = cmd.Process
	logger.Infof("Started Node.js server (PID: %d) on port %s", p.process.Pid, config.Port)
}

// proxyToNode forwards the original request to Node.js and returns the response.
func (p *NodeJSPlugin) proxyToNode(w http.ResponseWriter, r *http.Request, config NodeJSPluginConfig, logger *log.Logger) {
	// Construct the URL for forwarding the request to Node.js.
	nodeURL := fmt.Sprintf("http://localhost:%s%s", config.Port, r.URL.Path)

	// Create a new HTTP request forwarding the original request data.
	bodyReader, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Errorf("Failed to read request body: %v", err)
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	req, err := http.NewRequest(r.Method, nodeURL, strings.NewReader(string(bodyReader)))
	if err != nil {
		logger.Errorf("Failed to create request for Node.js: %v", err)
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	// Copy headers from the original request.
	for key, values := range r.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Errorf("Failed to connect to Node.js backend: %v", err)
		http.Error(w, "Node.js backend unavailable", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy response headers.
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Write the status code and response body.
	w.WriteHeader(resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Errorf("Failed to read response body from Node.js: %v", err)
		http.Error(w, "Failed to read response from Node.js", http.StatusInternalServerError)
		return
	}

	w.Write(body)
}

// startNodeServer ensures the Node.js server is running.
func (p *NodeJSPlugin) startNodeServer(config NodeJSPluginConfig, logger *log.Logger) {
	logger.Infof("Starting Node.js server...")
	p.mu.Lock()
	defer p.mu.Unlock()

	// Install dependencies if needed.
	if config.InstallDeps {
		p.installDependencies(config, logger)
	}

	// Start the Node.js server.
	entryPath := filepath.Join(config.RootDir, config.Entry)
	cmd := exec.Command(config.NodePath, entryPath)
	cmd.Dir = config.RootDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		logger.Errorf("Failed to start Node.js server: %v", err)
		return
	}

	p.process = cmd.Process
	logger.Infof("Started Node.js server (PID: %d) on port %s", p.process.Pid, config.Port)
}

// installDependencies installs dependencies using the configured package manager.
func (p *NodeJSPlugin) installDependencies(config NodeJSPluginConfig, logger *log.Logger) {
	nodeModulesPath := filepath.Join(config.RootDir, "node_modules")
	if _, err := os.Stat(nodeModulesPath); os.IsNotExist(err) {
		logger.Infof("node_modules not found, installing dependencies in %s", config.RootDir)

		packageManager := config.PackageManager
		if packageManager == "" {
			packageManager = "npm"
		}

		logger.Infof("Using package manager: %s", packageManager)
		cmd := exec.Command(packageManager, "install")
		cmd.Dir = config.RootDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			logger.Errorf("Failed to install dependencies using %s: %v", packageManager, err)
		} else {
			logger.Infof("Dependencies installed successfully using %s", packageManager)
		}
	}
}
