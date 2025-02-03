package plugin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mirkobrombin/goup/internal/config"
	log "github.com/sirupsen/logrus"
)

type MockPlugin struct{}

func (m *MockPlugin) Name() string  { return "MockPlugin" }
func (m *MockPlugin) OnInit() error { return nil }
func (m *MockPlugin) OnInitForSite(conf config.SiteConfig, logger *log.Logger) error {
	return nil
}
func (m *MockPlugin) BeforeRequest(r *http.Request) {}
func (m *MockPlugin) HandleRequest(w http.ResponseWriter, r *http.Request) bool {
	return false
}
func (m *MockPlugin) AfterRequest(w http.ResponseWriter, r *http.Request) {}
func (m *MockPlugin) OnExit() error                                       { return nil }

func TestPluginManager(t *testing.T) {
	pm := GetPluginManagerInstance()
	pm.Register(&MockPlugin{})

	if err := pm.InitPlugins(); err != nil {
		t.Fatalf("Failed to initialize plugins globally: %v", err)
	}

	conf := config.SiteConfig{
		Domain: "example.com",
	}
	logger := log.New()
	if err := pm.InitPluginsForSite(conf, logger); err != nil {
		t.Fatalf("Failed to initialize plugins for site: %v", err)
	}
}

func TestPluginMiddleware(t *testing.T) {
	pm := NewPluginManager()
	pm.Register(&MockPlugin{})

	mw := PluginMiddleware(pm)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "http://example.com", nil)
	rec := httptest.NewRecorder()

	mw.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}
