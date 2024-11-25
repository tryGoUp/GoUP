# GoUP! - A Minimal Configurable Web Server in Go

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
