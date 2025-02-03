package plugins

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/mirkobrombin/goup/internal/config"
	log "github.com/sirupsen/logrus"
)

// AuthPlugin provides HTTP Basic Authentication for protected paths.
type AuthPlugin struct {
	// Keeps the parsed config for the plugin.
	conf AuthPluginConfig
	// State holds the active sessions.
	state *AuthPluginState
	// Logger instance for this plugin.
	logger *log.Logger
}

// AuthPluginConfig represents the configuration for the AuthPlugin.
type AuthPluginConfig struct {
	// URL paths to protect with authentication.
	ProtectedPaths []string `json:"protected_paths"`
	// username:password pairs for authentication.
	Credentials map[string]string `json:"credentials"`
	// Session expiration in seconds.
	// -1 means sessions never expire. Maximum allowed is 86400 seconds (24 hours).
	SessionExpiration int `json:"session_expiration"`
}

// session represents an authenticated session.
type session struct {
	Username string
	Expiry   time.Time
}

// AuthPluginState internal state for the plugin.
type AuthPluginState struct {
	sessions map[string]session
	mu       sync.RWMutex
}

// Name returns the name of the plugin.
func (p *AuthPlugin) Name() string {
	return "AuthPlugin"
}

// OnInit is called once during the global plugin initialization.
func (p *AuthPlugin) OnInit() error {
	// No global setup needed here for now.
	return nil
}

// OnInitForSite is called for each site configuration.
func (p *AuthPlugin) OnInitForSite(conf config.SiteConfig, logger *log.Logger) error {
	p.logger = logger
	p.state = &AuthPluginState{
		sessions: make(map[string]session),
	}

	// Try to parse this plugin's config if present.
	pluginConfigRaw, ok := conf.PluginConfigs[p.Name()]
	if !ok {
		// If there's no AuthPlugin config, just do nothing.
		return nil
	}

	// Parse plugin configuration.
	authConfig := AuthPluginConfig{}
	if rawMap, ok := pluginConfigRaw.(map[string]interface{}); ok {
		// ProtectedPaths
		if paths, ok := rawMap["protected_paths"].([]interface{}); ok {
			for _, path := range paths {
				if pStr, ok := path.(string); ok {
					authConfig.ProtectedPaths = append(authConfig.ProtectedPaths, pStr)
				}
			}
		}

		// Credentials
		if creds, ok := rawMap["credentials"].(map[string]interface{}); ok {
			authConfig.Credentials = make(map[string]string)
			for user, pass := range creds {
				if passStr, ok := pass.(string); ok {
					authConfig.Credentials[user] = passStr
				}
			}
		}

		// SessionExpiration
		if se, ok := rawMap["session_expiration"].(float64); ok {
			authConfig.SessionExpiration = int(se)
		}
	}

	// Validate session expiration
	if authConfig.SessionExpiration > 86400 {
		return errors.New("session_expiration cannot exceed 86400 seconds (24 hours)")
	}
	if authConfig.SessionExpiration < -1 {
		return errors.New("session_expiration cannot be less than -1")
	}

	p.conf = authConfig

	// Initialization of the plugin state with optional session cleanup.
	if p.conf.SessionExpiration != -1 {
		go p.state.cleanupExpiredSessions(time.Minute, logger)
	}

	logger.Infof("Initializing AuthPlugin for domain: %s with session_expiration: %d",
		conf.Domain, authConfig.SessionExpiration)

	return nil
}

// BeforeRequest is invoked before serving each request.
func (p *AuthPlugin) BeforeRequest(r *http.Request) {
	// No specific pre-processing is needed here; logic is in HandleRequest.
}

// HandleRequest can fully handle the request, returning true if it does so.
func (p *AuthPlugin) HandleRequest(w http.ResponseWriter, r *http.Request) bool {
	// If there's no plugin config, do nothing.
	if p.conf.Credentials == nil {
		return false
	}

	// Check if the path is protected.
	protected := false
	for _, path := range p.conf.ProtectedPaths {
		if strings.HasPrefix(r.URL.Path, path) {
			protected = true
			break
		}
	}
	if !protected {
		// Not protected, continue with normal flow.
		return false
	}

	// The path is protected, check session or credentials.
	ip := getClientIP(r)
	if sess, exists := p.state.getSession(ip); exists {
		p.logger.Infof("Session valid for IP: %s, user: %s", ip, sess.Username)
		return false // Let the next handler continue.
	}

	// No valid session, check for Authorization header.
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		unauthorized(w)
		return true
	}

	// Parse Basic Auth
	username, password, ok := parseBasicAuth(authHeader)
	if !ok {
		unauthorized(w)
		return true
	}

	// Validate credentials
	expectedPassword, userExists := p.conf.Credentials[username]
	if !userExists || expectedPassword != password {
		unauthorized(w)
		return true
	}

	// Create a new session
	p.state.createSession(ip, username, p.conf.SessionExpiration, p.logger)
	p.logger.Infof("Authenticated IP: %s, user: %s", ip, username)

	// Return false to continue normal flow.
	return false
}

// AfterRequest is invoked after the request has been served or handled.
func (p *AuthPlugin) AfterRequest(w http.ResponseWriter, r *http.Request) {
	// Nothing to do here in this plugin.
}

// OnExit is called when the server is shutting down.
func (p *AuthPlugin) OnExit() error {
	// No cleanup needed, but you could stop session cleanup goroutines if needed.
	return nil
}

// getClientIP extracts the client's IP address from the request.
func getClientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	if ips := r.Header.Get("X-Forwarded-For"); ips != "" {
		// X-Forwarded-For may contain multiple IPs, take the first one
		return strings.Split(ips, ",")[0]
	}
	// Fallback to RemoteAddr
	ip := r.RemoteAddr
	if colonIndex := strings.LastIndex(ip, ":"); colonIndex != -1 {
		ip = ip[:colonIndex]
	}
	return ip
}

// parseBasicAuth parses the Basic Authentication header.
func parseBasicAuth(authHeader string) (username, password string, ok bool) {
	const prefix = "Basic "
	if !strings.HasPrefix(authHeader, prefix) {
		return
	}
	decoded, err := base64.StdEncoding.DecodeString(authHeader[len(prefix):])
	if err != nil {
		return
	}
	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return
	}
	username = parts[0]
	password = parts[1]
	ok = true
	return
}

// unauthorized sends a 401 Unauthorized response with the appropriate header.
func unauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}

// getSession retrieves a session for the given IP, if it exists and is valid.
func (s *AuthPluginState) getSession(ip string) (session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sess, exists := s.sessions[ip]
	if !exists {
		return session{}, false
	}

	// Check expiration
	if !sess.Expiry.IsZero() && sess.Expiry.Before(time.Now()) {
		return session{}, false
	}
	return sess, true
}

// createSession creates a new session for the given IP and username.
func (s *AuthPluginState) createSession(ip, username string, expiration int, logger *log.Logger) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var expiry time.Time
	if expiration != -1 {
		expiry = time.Now().Add(time.Duration(expiration) * time.Second)
	}
	s.sessions[ip] = session{
		Username: username,
		Expiry:   expiry,
	}
	if expiration != -1 {
		logger.Infof("Session created for IP: %s, user: %s, expires at: %v", ip, username, expiry)
	} else {
		logger.Infof("Session created for IP: %s, user: %s, never expires", ip, username)
	}
}

// cleanupExpiredSessions periodically removes expired sessions.
func (s *AuthPluginState) cleanupExpiredSessions(interval time.Duration, logger *log.Logger) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		s.mu.Lock()
		for ip, sess := range s.sessions {
			if !sess.Expiry.IsZero() && sess.Expiry.Before(time.Now()) {
				delete(s.sessions, ip)
				logger.Infof("Session expired and removed for IP: %s, user: %s", ip, sess.Username)
			}
		}
		s.mu.Unlock()
	}
}
