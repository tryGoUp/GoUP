package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/mirkobrombin/goup/internal/config"
	log "github.com/sirupsen/logrus"
)

// NewLogger sets up a new logger for the given identifier.
func NewLogger(identifier string) (*log.Logger, error) {
	logger := log.New()

	// Log directory: ~/.local/share/goup/logs/identifier/year/month/
	logDir := filepath.Join(config.GetLogDir(), identifier, fmt.Sprintf("%d", time.Now().Year()), fmt.Sprintf("%02d", time.Now().Month()))
	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		return nil, err
	}

	// Log name format: day.log
	logFile := filepath.Join(logDir, fmt.Sprintf("%02d.log", time.Now().Day()))
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	// Setting up output in both stdout and file
	logger.SetOutput(io.MultiWriter(os.Stdout, file))
	logger.SetFormatter(&log.JSONFormatter{})

	return logger, nil
}
