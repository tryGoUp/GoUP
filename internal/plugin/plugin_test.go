package plugin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mirkobrombin/goup/internal/server/middleware"
	"github.com/mirkobrombin/goup/plugins"
)

func TestPluginManager(t *testing.T) {
	pluginManager := GetPluginManagerInstance()

	pluginManager.Register(&plugins.CustomHeaderPlugin{})

	mwManager := middleware.NewMiddlewareManager()
	err := pluginManager.InitPlugins(mwManager)
	if err != nil {
		t.Fatalf("Failed to initialize plugin: %v", err)
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
