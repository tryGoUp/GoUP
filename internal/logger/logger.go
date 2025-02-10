package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/mirkobrombin/goup/internal/config"
	"github.com/rs/zerolog"
)

// Fields is a map of string keys to arbitrary values, emulating logrus.Fields
// for compatibility with existing code.
type Fields map[string]interface{}

// Logger wraps a zerolog.Logger while exposing methods similar to logrus.
type Logger struct {
	base zerolog.Logger
	out  io.Writer
}

// SetOutput changes the output writer (stdout, file, etc.).
func (l *Logger) SetOutput(w io.Writer) {
	l.out = w
	l.base = l.base.Output(w)
}

// WithFields returns a new Logger that includes the provided fields.
func (l *Logger) WithFields(fields Fields) *Logger {
	newBase := l.base.With().Fields(fields).Logger()
	return &Logger{
		base: newBase,
		out:  l.out,
	}
}

// Info logs a message at Info level.
func (l *Logger) Info(msg string) {
	l.base.Info().Msg(msg)
}

// Infof logs a formatted message at Info level.
func (l *Logger) Infof(format string, args ...interface{}) {
	l.base.Info().Msgf(format, args...)
}

// Error logs a message at Error level.
func (l *Logger) Error(msg string) {
	l.base.Error().Msg(msg)
}

// Errorf logs a formatted message at Error level.
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.base.Error().Msgf(format, args...)
}

// Debug logs a message at Debug level (not heavily used by default).
func (l *Logger) Debug(msg string) {
	l.base.Debug().Msg(msg)
}

// Debugf logs a formatted message at Debug level.
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.base.Debug().Msgf(format, args...)
}

// Warn logs a message at Warn level.
func (l *Logger) Warn(msg string) {
	l.base.Warn().Msg(msg)
}

// Warnf logs a formatted message at Warn level.
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.base.Warn().Msgf(format, args...)
}

// NewLogger creates a new Logger that writes both to stdout and a site-specific file.
func NewLogger(identifier string, fields Fields) (*Logger, error) {
	logDir := filepath.Join(
		config.GetLogDir(),
		identifier,
		fmt.Sprintf("%d", time.Now().Year()),
		fmt.Sprintf("%02d", time.Now().Month()),
	)
	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		return nil, err
	}

	// Log file name format: 03.log (day.log)
	logFile := filepath.Join(logDir, fmt.Sprintf("%02d.log", time.Now().Day()))
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	mw := io.MultiWriter(os.Stdout, file)

	// Zerolog logger with time + multiwriter
	base := zerolog.New(mw).With().Timestamp().Logger()

	l := &Logger{
		base: base,
		out:  mw,
	}

	if fields != nil {
		l = l.WithFields(fields)
	}

	return l, nil
}

// NewPluginLogger creates a plugin-specific log file (no stdout).
func NewPluginLogger(siteDomain, pluginName string) (*Logger, error) {
	logDir := filepath.Join(
		config.GetLogDir(),
		siteDomain,
		fmt.Sprintf("%d", time.Now().Year()),
		fmt.Sprintf("%02d", time.Now().Month()),
	)
	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		return nil, err
	}

	// Log file format: 03-NomePlugin.log (day-PluginName.log)
	logFile := filepath.Join(logDir, fmt.Sprintf("%02d-%s.log", time.Now().Day(), pluginName))
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	// Only file output with timestamp
	base := zerolog.New(file).With().Timestamp().Logger()

	l := &Logger{
		base: base,
		out:  file,
	}
	return l, nil
}

// Writer returns an io.WriteCloser that logs each written line.
func (l *Logger) Writer() io.WriteCloser {
	pr, pw := io.Pipe()

	go func() {
		defer pr.Close()
		buf := make([]byte, 1024)
		var tmp []byte

		for {
			n, err := pr.Read(buf)
			if n > 0 {
				tmp = append(tmp, buf[:n]...)
				for {
					idx := indexOfNewline(tmp)
					if idx == -1 {
						break
					}
					line := tmp[:idx]
					line = trimCR(line)
					l.Info(string(line))
					tmp = tmp[idx+1:]
				}
			}
			if err != nil {
				// Exit on error or EOF
				break
			}
		}
		// Logging any remaining data
		if len(tmp) > 0 {
			l.Info(string(tmp))
		}
	}()

	return &pipeWriteCloser{
		pipeWriter: pw,
	}
}

// pipeWriteCloser implements Write and Close delegating to a PipeWriter.
type pipeWriteCloser struct {
	pipeWriter *io.PipeWriter
}

func (pwc *pipeWriteCloser) Write(data []byte) (int, error) {
	return pwc.pipeWriter.Write(data)
}

func (pwc *pipeWriteCloser) Close() error {
	return pwc.pipeWriter.Close()
}

func indexOfNewline(buf []byte) int {
	for i, b := range buf {
		if b == '\n' {
			return i
		}
	}
	return -1
}

func trimCR(buf []byte) []byte {
	if len(buf) > 0 && buf[len(buf)-1] == '\r' {
		return buf[:len(buf)-1]
	}
	return buf
}
