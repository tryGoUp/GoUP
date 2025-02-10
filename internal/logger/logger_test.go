package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mirkobrombin/goup/internal/config"
)

func TestNewLogger(t *testing.T) {
	identifier := "test_identifier"

	tmpDir := t.TempDir()
	config.SetCustomLogDir(tmpDir)

	fields := Fields{
		"test_field": "test_value",
	}

	logger, err := NewLogger(identifier, fields)
	if err != nil {
		t.Fatalf("Error creating new logger: %v", err)
	}
	if logger == nil {
		t.Fatalf("Logger is nil")
	}

	year := time.Now().Year()
	month := time.Now().Month()
	day := time.Now().Day()

	logDir := filepath.Join(tmpDir, identifier, fmt.Sprintf("%d", year), fmt.Sprintf("%02d", month))
	logFile := filepath.Join(logDir, fmt.Sprintf("%02d.log", day))

	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Errorf("Expected log file %s to exist", logFile)
	}

	// Just a quick test of logging methods
	logger.Infof("Testing Infof %s", "works")
	logger.Errorf("Testing Errorf %s", "works")
}

func TestNewPluginLogger(t *testing.T) {
	siteDomain := "example-plugin"
	pluginName := "testPlugin"

	tmpDir := t.TempDir()
	config.SetCustomLogDir(tmpDir)

	logger, err := NewPluginLogger(siteDomain, pluginName)
	if err != nil {
		t.Fatalf("Error creating plugin logger: %v", err)
	}
	if logger == nil {
		t.Fatalf("Plugin logger is nil")
	}

	year := time.Now().Year()
	month := time.Now().Month()
	day := time.Now().Day()

	logDir := filepath.Join(tmpDir, siteDomain, fmt.Sprintf("%d", year), fmt.Sprintf("%02d", month))
	logFile := filepath.Join(logDir, fmt.Sprintf("%02d-%s.log", day, pluginName))

	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Errorf("Expected plugin log file %s to exist", logFile)
	}

	logger.Info("Some plugin log message")
}

// TestLoggerWriter checks if Logger.Writer() properly logs lines.
func TestLoggerWriter(t *testing.T) {
	tmpDir := t.TempDir()
	config.SetCustomLogDir(tmpDir)

	logger, err := NewLogger("writer_test", nil)
	if err != nil {
		t.Fatalf("Error creating logger: %v", err)
	}
	if logger == nil {
		t.Fatalf("Logger is nil")
	}

	// Prepare the log file path
	year := time.Now().Year()
	month := time.Now().Month()
	day := time.Now().Day()
	logDir := filepath.Join(tmpDir, "writer_test", fmt.Sprintf("%d", year), fmt.Sprintf("%02d", month))
	logFile := filepath.Join(logDir, fmt.Sprintf("%02d.log", day))

	// Write directly into Logger.Writer()
	w := logger.Writer()
	lines := []string{
		"Hello from Writer() test",
		"Another line from Writer()",
	}
	// Each line ends with \n so that the logger processes them
	for _, line := range lines {
		io.WriteString(w, line+"\n")
	}
	// Give some time to the goroutine that processes lines
	time.Sleep(200 * time.Millisecond)

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Error reading log file: %v", err)
	}
	content := string(data)

	for _, line := range lines {
		if !strings.Contains(content, line) {
			t.Errorf("Expected log file to contain: %q, but it was not found\nLog content:\n%s", line, content)
		}
	}
}
