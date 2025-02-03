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
	"github.com/mirkobrombin/goup/internal/logger"
	log "github.com/sirupsen/logrus"
)

// NodeJSPlugin handles the execution of a Node.js application.
type NodeJSPlugin struct {
	mu     sync.Mutex
	logger *log.Logger
	// A reference to the currently running Node.js process (if any).
	process *os.Process
	// siteConfigs holds domain-specific NodeJSPluginConfig.
	siteConfigs map[string]NodeJSPluginConfig
}

// Name returns the name of the plugin.
func (p *NodeJSPlugin) Name() string {
	return "NodeJSPlugin"
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

// OnInit registers any global plugin logic (none in this case).
func (p *NodeJSPlugin) OnInit() error {
	p.siteConfigs = make(map[string]NodeJSPluginConfig)
	return nil
}

// OnInitForSite initializes the plugin for a specific site.
func (p *NodeJSPlugin) OnInitForSite(conf config.SiteConfig, baseLogger *log.Logger) error {
	if p.logger == nil {
		// Create a dedicated logger for NodeJSPlugin overall.
		pluginLogger, err := logger.NewPluginLogger(conf.Domain, p.Name())
		if err != nil {
			baseLogger.Errorf("Failed to create NodeJSPlugin logger: %v", err)
			return err
		}
		p.logger = pluginLogger
	}

	// Parse NodeJSPluginConfig from the site config.
	pluginConfigRaw, ok := conf.PluginConfigs[p.Name()]
	if ok {
		cfg := NodeJSPluginConfig{}
		if rawMap, ok := pluginConfigRaw.(map[string]interface{}); ok {
			if enable, ok := rawMap["enable"].(bool); ok {
				cfg.Enable = enable
			}
			if port, ok := rawMap["port"].(string); ok {
				cfg.Port = port
			}
			if rootDir, ok := rawMap["root_dir"].(string); ok {
				cfg.RootDir = rootDir
			}
			if entry, ok := rawMap["entry"].(string); ok {
				cfg.Entry = entry
			}
			if installDeps, ok := rawMap["install_deps"].(bool); ok {
				cfg.InstallDeps = installDeps
			}
			if nodePath, ok := rawMap["node_path"].(string); ok {
				cfg.NodePath = nodePath
			}
			if packageManager, ok := rawMap["package_manager"].(string); ok {
				cfg.PackageManager = packageManager
			}
			if proxyPaths, ok := rawMap["proxy_paths"].([]interface{}); ok {
				for _, pathVal := range proxyPaths {
					if pathStr, ok := pathVal.(string); ok {
						cfg.ProxyPaths = append(cfg.ProxyPaths, pathStr)
					}
				}
			}
		}
		p.siteConfigs[conf.Domain] = cfg
	} else {
		// If there's no config for this domain, store a default disabled config.
		p.siteConfigs[conf.Domain] = NodeJSPluginConfig{}
	}

	return nil
}

// BeforeRequest is invoked before serving each request (unused here).
func (p *NodeJSPlugin) BeforeRequest(r *http.Request) {}

// HandleRequest can fully handle the request, returning true if it does so.
func (p *NodeJSPlugin) HandleRequest(w http.ResponseWriter, r *http.Request) bool {
	// Identify the domain and strip any port.
	host := r.Host
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}

	cfg, ok := p.siteConfigs[host]
	if !ok || !cfg.Enable {
		return false
	}

	// Ensure Node.js is running if needed.
	p.ensureNodeServerRunning(cfg)

	// Check if path matches one of the ProxyPaths.
	for _, proxyPath := range cfg.ProxyPaths {
		if strings.HasPrefix(r.URL.Path, proxyPath) {
			p.proxyToNode(w, r, cfg)
			return true
		}
	}

	// Not handled, continue with other handlers.
	return false
}

// AfterRequest is invoked after the request has been served or handled.
func (p *NodeJSPlugin) AfterRequest(w http.ResponseWriter, r *http.Request) {}

// OnExit is called when the server is shutting down.
func (p *NodeJSPlugin) OnExit() error {
	// Optionally, terminate Node.js if running.
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.process != nil {
		p.logger.Infof("Terminating Node.js process (PID: %d).", p.process.Pid)
		_ = p.process.Kill()
		p.process = nil
	}
	return nil
}

// ensureNodeServerRunning starts Node.js if it is not already running.
func (p *NodeJSPlugin) ensureNodeServerRunning(config NodeJSPluginConfig) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.process != nil {
		// Already running, do nothing.
		return
	}

	p.logger.Infof("Starting Node.js server...")

	// Install dependencies if required.
	if config.InstallDeps {
		p.installDependencies(config)
	}

	// Start the Node.js server.
	entryPath := filepath.Join(config.RootDir, config.Entry)
	nodePath := config.NodePath
	if nodePath == "" {
		nodePath = "node"
	}

	cmd := exec.Command(nodePath, entryPath)
	cmd.Dir = config.RootDir
	cmd.Stdout = p.logger.Writer()
	cmd.Stderr = p.logger.Writer()

	if err := cmd.Start(); err != nil {
		p.logger.Errorf("Failed to start Node.js server: %v", err)
		return
	}

	p.process = cmd.Process
	p.logger.Infof("Started Node.js server (PID: %d) on port %s", p.process.Pid, config.Port)

	// Watch for process exit.
	go func() {
		err := cmd.Wait()
		p.logger.Infof("Node.js server exited (PID: %d), error=%v", p.process.Pid, err)
		p.logger.Writer().Close()
	}()
}

// proxyToNode forwards the request to Node.js and sends back the response.
func (p *NodeJSPlugin) proxyToNode(w http.ResponseWriter, r *http.Request, config NodeJSPluginConfig) {
	nodeURL := fmt.Sprintf("http://localhost:%s%s", config.Port, r.URL.Path)
	bodyReader, err := io.ReadAll(r.Body)
	if err != nil {
		p.logger.Errorf("Failed to read request body: %v", err)
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	req, err := http.NewRequest(r.Method, nodeURL, strings.NewReader(string(bodyReader)))
	if err != nil {
		p.logger.Errorf("Failed to create request for Node.js: %v", err)
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	// Copy headers
	for key, values := range r.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		p.logger.Errorf("Failed to connect to Node.js backend: %v", err)
		http.Error(w, "Node.js backend unavailable", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Forward response headers.
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		p.logger.Errorf("Failed to read response from Node.js: %v", err)
		http.Error(w, "Failed to read response from Node.js", http.StatusInternalServerError)
		return
	}
	w.Write(body)
}

// installDependencies installs dependencies using the configured package manager.
func (p *NodeJSPlugin) installDependencies(config NodeJSPluginConfig) {
	nodeModulesPath := filepath.Join(config.RootDir, "node_modules")
	if _, err := os.Stat(nodeModulesPath); os.IsNotExist(err) {
		p.logger.Infof("node_modules not found, installing dependencies in %s", config.RootDir)
		packageManager := config.PackageManager
		if packageManager == "" {
			packageManager = "npm"
		}
		cmd := exec.Command(packageManager, "install")
		cmd.Dir = config.RootDir
		cmd.Stdout = p.logger.Writer()
		cmd.Stderr = p.logger.Writer()

		if err := cmd.Run(); err != nil {
			p.logger.Errorf("Failed to install dependencies using %s: %v", packageManager, err)
		} else {
			p.logger.Infof("Dependencies installed successfully using %s", packageManager)
		}
	}
}
