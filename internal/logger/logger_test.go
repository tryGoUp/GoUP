package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mirkobrombin/goup/internal/config"
)

func TestNewLogger(t *testing.T) {
	identifier := "test_identifier"

	tmpDir := t.TempDir()
	config.SetCustomLogDir(tmpDir)
	logger, err := NewLogger(identifier)
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
}
