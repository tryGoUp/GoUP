package plugin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/mirkobrombin/goup/internal/config"
	"github.com/mirkobrombin/goup/internal/server/middleware"
	"github.com/mirkobrombin/goup/plugins"
)

func TestPluginManager(t *testing.T) {
	pluginManager := GetPluginManagerInstance()
	pluginManager.Register(&plugins.CustomHeaderPlugin{})
	mwManager := middleware.NewMiddlewareManager()

	conf := config.SiteConfig{
		CustomHeaders: map[string]string{
			"X-GoUP-Header": "GoUP",
		},
	}

	logger := log.New()
	err := pluginManager.InitPluginsForSite(mwManager, logger, conf)
	if err != nil {
		t.Fatalf("Failed to initialize plugin for site: %v", err)
	}

	handler := mwManager.Apply(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "http://example.com", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Header().Get("X-GoUP-Header") != "GoUP" {
		t.Errorf("Expected X-GoUP-Header to be 'GoUP', got '%s'", rec.Header().Get("X-GoUP-Header"))
	}
}
