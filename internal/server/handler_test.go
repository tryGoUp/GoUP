package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/mirkobrombin/goup/internal/config"
	"github.com/mirkobrombin/goup/internal/server/middleware"
	log "github.com/sirupsen/logrus"
)

func TestCreateHandler_Static(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_static")
	if err != nil {
		t.Fatalf("Error creating temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFilePath := filepath.Join(tmpDir, "testfile.txt")
	err = os.WriteFile(testFilePath, []byte("Test content"), 0644)
	if err != nil {
		t.Fatalf("Error creating test file: %v", err)
	}

	conf := config.SiteConfig{
		Domain:        "example.com",
		RootDirectory: tmpDir,
		CustomHeaders: map[string]string{
			"X-Test-Header": "TestValue",
		},
		RequestTimeout: 60,
	}
	logger := log.New()
	logger.Out = os.Stderr
	identifier := "test"

	handler, err := createHandler(conf, logger, identifier, &middleware.MiddlewareManager{})
	if err != nil {
		t.Fatalf("Error creating handler: %v", err)
	}

	req := httptest.NewRequest("GET", "http://example.com/testfile.txt", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	if w.Header().Get("X-Test-Header") != "TestValue" {
		t.Errorf("Expected custom header 'X-Test-Header' to be 'TestValue', got %q", w.Header().Get("X-Test-Header"))
	}

	expectedBody := "Test content"
	if w.Body.String() != expectedBody {
		t.Errorf("Expected body %q, got %q", expectedBody, w.Body.String())
	}
}

func TestCreateHandler_ProxyPass(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Backend Response"))
	}))
	defer backend.Close()

	conf := config.SiteConfig{
		Domain:         "example.com",
		ProxyPass:      backend.URL,
		CustomHeaders:  map[string]string{},
		RequestTimeout: 60,
	}
	logger := log.New()
	logger.Out = os.Stderr
	identifier := "test"

	handler, err := createHandler(conf, logger, identifier, &middleware.MiddlewareManager{})
	if err != nil {
		t.Fatalf("Error creating handler: %v", err)
	}

	req := httptest.NewRequest("GET", "http://example.com/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	expectedBody := "Backend Response"
	if w.Body.String() != expectedBody {
		t.Errorf("Expected body %q, got %q", expectedBody, w.Body.String())
	}
}
