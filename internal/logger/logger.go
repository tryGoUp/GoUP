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

// FieldHook is a custom Logrus hook that adds predefined fields to every log entry.
type FieldHook struct {
	Fields log.Fields
}

// Levels defines the log levels where the hook is applied.
func (hook *FieldHook) Levels() []log.Level {
	return log.AllLevels
}

// Fire adds the predefined fields to the log entry.
func (hook *FieldHook) Fire(entry *log.Entry) error {
	for k, v := range hook.Fields {
		entry.Data[k] = v
	}
	return nil
}

// NewLogger creates a new logger with optional predefined fields.
func NewLogger(identifier string, fields log.Fields) (*log.Logger, error) {
	logger := log.New()

	// Directory for logs: ~/.local/share/goup/logs/identifier/year/month/
	logDir := filepath.Join(config.GetLogDir(), identifier, fmt.Sprintf("%d", time.Now().Year()), fmt.Sprintf("%02d", time.Now().Month()))
	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		return nil, err
	}

	// Log file name: day.log
	logFile := filepath.Join(logDir, fmt.Sprintf("%02d.log", time.Now().Day()))
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	// Set output to both stdout and the log file
	logger.SetOutput(io.MultiWriter(os.Stdout, file))
	logger.SetFormatter(&log.JSONFormatter{})

	// Add the FieldHook if fields are provided
	if fields != nil {
		logger.AddHook(&FieldHook{Fields: fields})
	}

	return logger, nil
}
