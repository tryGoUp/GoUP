<div align="center">
    <img src="https://raw.githubusercontent.com/tryGoUp/brand/refs/heads/main/Logo-Text/Light/logo-text-light.png#gh-light-mode-only" height="100">
    <img src="https://raw.githubusercontent.com/tryGoUp/brand/refs/heads/main/Logo-Text/Dark/logo-text-dark.png#gh-dark-mode-only" height="100">
    <hr />
    <p>A Minimal Configurable Web Server in Go</p>
</div>

GoUP! is a minimal, tweakable web server written in Go. You can use it to serve static files, set up reverse proxies, and configure SSL for multiple domains, all through simple JSON configuration files. GoUp spawns a dedicated server for each port, websites with the same port are treated as virtual hosts and run on the same server.

## Features

- Serve static files from a specified root directory
- Set up reverse proxies to backend services
- Support for SSL/TLS with custom certificates
- Custom headers for HTTP responses
- Support for multiple domains and virtual hosting
- Logging to both console and files - JSON formatted (structured logs)
- Optional TUI interface for real-time monitoring
- HTTP/2 and HTTP/3 support (not configurable, HTTP/1.1 is used for unencrypted connections, HTTP/2 and HTTP/3 for encrypted connections)

## Future Plans

- API for dynamic configuration changes
- Docker/Podman support for easy deployment

## Installation

Go is required to build the software, ensure you have it installed on your system.

1. **Clone the repository:**

   ```bash
   git clone https://github.com/mirkobrombin/goup.git
   cd goup
   ```

2. **Build the software:**

   ```bash
   go build -o ~/.local/bin/goup cmd/goup/main.go
   ```


## Usage

### Generating a Configuration

The `generate` command can be used to create a new site configuration:

```bash
goup generate [domain]
```

Example:

```bash
goup generate example.com
```

it will prompt you to enter the following details:

- Port number
- Root directory
- Proxy settings
- Custom headers
- SSL configuration
- Request timeout

The configuration file will be saved in `~/.config/goup/` as `[domain].json`, you
can edit or create new configurations manually, just restart the server to apply changes.

Read more about the configuration structure in the [Configuration](#configuration) section.

### Starting the Server

Start the server with:

```bash
goup start
```

This command loads all configurations from `~/.config/goup/` and starts serving.

**Starting with TUI Mode:**

Enable the Text-Based User Interface to monitor logs:

```bash
goup start --tui
```

### Additional Commands

- **List Configured Sites:**

  ```bash
  goup list
  ```

- **Validate Configurations:**

  ```bash
  goup validate
  ```

- **Stop the Server:**

  ```bash
  goup stop // Not implemented yet, use <Ctrl+C> to stop the server
  ```

- **Restart the Server:**

  ```bash
  goup restart // Not implemented yet, use <Ctrl+C> to stop the server and start it again
  ```

## Configuration

### Site Configuration Structure

Each site configuration is represented by a JSON file and meets the following structure:

```json
{
  "domain": "example.com",
  "port": 8080,
  "root_directory": "/path/to/root",
  "custom_headers": {
      "X-Domain-Name": "example.com"
  },
  "proxy_pass": "http://localhost:3000",
  "ssl": {
    "enabled": true,
    "certificate": "/path/to/cert.crt",
    "key": "/path/to/key.key"
  },
  "request_timeout": 60
}
```

**Fields:**

- **domain**: The domain name for which the server will respond
- **port**: The port number to listen on
- **root_directory**: Path to the directory containing static files. Leave empty if using `proxy_pass`
- **custom_headers**: Key-value pairs of custom headers to include in responses
- **proxy_pass**: URL to the backend service for reverse proxying. Leave empty if serving static files
- **ssl**:
  - **enabled**: Set to `true` to enable SSL/TLS
  - **certificate**: Path to the SSL certificate file
  - **key**: Path to the SSL key file
- **request_timeout**: Timeout for client requests in seconds

## Logging

Logs are written to both the console and log files, those are stored in:

```
~/.local/share/goup/logs/[identifier]/[year]/[month]/[day].log
```

- **identifier**: The domain name or `port_[port_number]` for virtual hosts
- Logs are formatted in JSON for easy parsing

## TUI Interface

Enable the TUI with the `--tui` flag when starting the server:

```bash
goup start --tui
```

## Plugins

GoUP! has a lightweight plugin system that allows you to extend its 
functionality. Plugins implement a set of hooks for initialization, request 
handling, and cleanup:

- **OnInit()**: Called once when GoUP! starts (useful for global setup).
- **OnInitForSite(conf config.SiteConfig, logger *log.Logger)**: Called for 
each site configuration (site-specific setup).
- **BeforeRequest(r *http.Request)**: Invoked before every request, letting 
you examine or modify the incoming request.
- **HandleRequest(w http.ResponseWriter, r *http.Request) bool**: If your 
plugin wants to fully handle the request (e.g., returning a response on its 
own), do it here. Return `true` if the request was fully handled (so GoUP! 
won’t process it further).
- **AfterRequest(w http.ResponseWriter, r *http.Request)**: Called after each 
request has been served or intercepted by this plugin.
- **OnExit()**: Called once when GoUP! is shutting down (for cleanup).

### Enabling Plugins

To enable plugins, add their configuration in the `plugin_configs` section of 
the site’s JSON configuration file. For example:

```json
{
  "domain": "example.com",
  "port": 8080,
  "root_directory": "/path/to/root",
  "custom_headers": {
    "X-Custom-Header": "Hello, World!"
  },
  "plugin_configs": {
    "PHPPlugin": {
      "enable": true,
      "fpm_addr": "/run/php/php8.2-fpm.sock"
    },
    "AuthPlugin": {
      "protected_paths": ["/protected.html"],
      "credentials": {
        "admin": "password123",
        "user": "userpass"
      },
      "session_expiration": 3600
    },
    "NodeJSPlugin": {
      "enable": true,
      "port": "3000",
      "root_dir": "/path/to/node/app",
      "entry": "server.js",
      "install_deps": true,
      "node_path": "/usr/bin/node",
      "package_manager": "pnpm",
      "proxy_paths": ["/api/", "/backend/"]
    }
  }
}
```

### Pre-Installed Plugins

- **Custom Header Plugin**: Adds custom headers to HTTP responses, configured 
per domain.
- **PHP Plugin**: Handles `.php` requests using PHP-FPM.
- **Auth Plugin**: Protects routes with basic authentication.
- **NodeJS Plugin**: Handles Node.js applications using `node`.

Each plugin can have its own JSON configuration under `plugin_configs`, which it 
reads in `OnInitForSite`.

### Developing Plugins

You can create your own plugins by implementing the following interface:

```go
type Plugin interface {
    Name() string
    OnInit() error
    OnInitForSite(conf config.SiteConfig, logger *log.Logger) error
    BeforeRequest(r *http.Request)
    HandleRequest(w http.ResponseWriter, r *http.Request) bool
    AfterRequest(w http.ResponseWriter, r *http.Request)
    OnExit() error
}
```

A minimal example plugin:

```go
package myplugin

import (
    "net/http"

    "github.com/mirkobrombin/goup/internal/config"
    log "github.com/sirupsen/logrus"
)

type MyPlugin struct{}

// Name returns the plugin's name.
func (p *MyPlugin) Name() string {
    return "MyPlugin"
}

// OnInit is called once globally on startup.
func (p *MyPlugin) OnInit() error {
    // Perform any global setup here.
    return nil
}

// OnInitForSite is called for each site configuration.
func (p *MyPlugin) OnInitForSite(conf config.SiteConfig, logger *log.Logger) error {
    // Site-specific setup (e.g. reading plugin config).
    logger.Infof("MyPlugin initialized for site: %s", conf.Domain)
    return nil
}

// BeforeRequest is invoked before handling every request.
func (p *MyPlugin) BeforeRequest(r *http.Request) {
    // Optionally inspect or modify the request.
}

// HandleRequest can take over the request entirely if desired.
// Return true if you want to finalize the response here.
func (p *MyPlugin) HandleRequest(w http.ResponseWriter, r *http.Request) bool {
    // Example: Just log and let GoUP! continue.
    return false
}

// AfterRequest is invoked after the request is served (or handled by this plugin).
func (p *MyPlugin) AfterRequest(w http.ResponseWriter, r *http.Request) {
    // Any post-processing goes here.
}

// OnExit is called when GoUP! is shutting down.
func (p *MyPlugin) OnExit() error {
    // Cleanup resources if needed.
    return nil
}
```

Then register your plugin in the `main.go` file:

```go
pluginManager.Register(&myplugin.MyPlugin{})
```

## Contributing

I really appreciate any contributions you would like to make, whether it's a 
simple typo fix or a new feature. Feel free to open an issue or submit a pull request.

## Pro Tips

You can use the `public/` directory in the repository as the root directory for
your test sites. It contains a simple `index.html` file with a JS script that
gets the website's title from the `X-Domain-Name` header (if set).

## License

GoUP! is released under the [MIT License](LICENSE).

---

**Note:** This project is for educational purposes and may not be suitable 
for production environments without additional security and performance 
considerations, yet.
