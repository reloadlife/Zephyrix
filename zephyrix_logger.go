package zephyrix

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

// String returns the string representation of a LogLevel
func (l LogLevel) String() string {
	return [...]string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}[l]
}

// LogConfig holds the configuration for the logger
type LogConfig struct {
	Level     string   `mapstructure:"level"`
	File      string   `mapstructure:"file"`
	Database  string   `mapstructure:"database"`
	RemoteURL string   `mapstructure:"remote_url"`
	Outputs   []string `mapstructure:"outputs"`
}

// ZephyrixLogger interface defines the methods that a logger should implement
type ZephyrixLogger interface {
	Debug(format string, v ...interface{})
	Info(format string, v ...interface{})
	Warn(format string, v ...interface{})
	Error(format string, v ...interface{})
	Fatal(format string, v ...interface{})
	SetLevel(level LogLevel)
	AddOutput(writer io.Writer)
}

// zephyrixLogger is the concrete implementation of the ZephyrixLogger interface
type zephyrixLogger struct {
	level   LogLevel
	outputs []io.Writer
	mu      sync.Mutex
}

// NewLogger creates a new ZephyrixLogger instance
func NewLogger(config LogConfig) (ZephyrixLogger, error) {
	level, err := parseLogLevel(config.Level)
	if err != nil {
		return nil, err
	}

	logger := &zephyrixLogger{
		level:   level,
		outputs: []io.Writer{os.Stdout}, // Default output
	}

	if err := logger.configureOutputs(config); err != nil {
		return nil, err
	}

	return logger, nil
}

func (z *zephyrix) setupLogger() error {
	newLogger, err := NewLogger(z.config.Log)
	if err != nil {
		return fmt.Errorf("failed to setup logger: %w", err)
	}
	Logger = newLogger
	return nil
}

func parseLogLevel(level string) (LogLevel, error) {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return LevelDebug, nil
	case "INFO":
		return LevelInfo, nil
	case "WARN":
		return LevelWarn, nil
	case "ERROR":
		return LevelError, nil
	case "FATAL":
		return LevelFatal, nil
	default:
		return LevelInfo, fmt.Errorf("invalid log level: %s", level)
	}
}

func (l *zephyrixLogger) configureOutputs(config LogConfig) error {
	for _, output := range config.Outputs {
		switch output {
		case "file":
			if config.File != "" {
				file, err := l.openLogFile(config.File)
				if err != nil {
					return err
				}
				l.AddOutput(file)
			}
		case "database":
			if config.Database != "" {
				// TODO: Implement database logging
				return fmt.Errorf("database logging not yet implemented")
			}
		case "remote":
			if config.RemoteURL != "" {
				// TODO: Implement remote logging
				return fmt.Errorf("remote logging not yet implemented")
			}
		}
	}
	return nil
}

func (l *zephyrixLogger) openLogFile(filename string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}
	return os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
}

func (l *zephyrixLogger) log(level LogLevel, format string, v ...interface{}) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	message := fmt.Sprintf(format, v...)
	logEntry := fmt.Sprintf("[%s] %s %s\n", time.Now().Format(time.RFC3339), level, message)

	for _, output := range l.outputs {
		_, _ = fmt.Fprint(output, logEntry)
	}

	if level == LevelFatal {
		os.Exit(1)
	}
}

func (l *zephyrixLogger) Debug(format string, v ...interface{}) { l.log(LevelDebug, format, v...) }
func (l *zephyrixLogger) Info(format string, v ...interface{})  { l.log(LevelInfo, format, v...) }
func (l *zephyrixLogger) Warn(format string, v ...interface{})  { l.log(LevelWarn, format, v...) }
func (l *zephyrixLogger) Error(format string, v ...interface{}) { l.log(LevelError, format, v...) }
func (l *zephyrixLogger) Fatal(format string, v ...interface{}) { l.log(LevelFatal, format, v...) }

func (l *zephyrixLogger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

func (l *zephyrixLogger) AddOutput(writer io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.outputs = append(l.outputs, writer)
}

// Logger is the global logger instance
var Logger ZephyrixLogger

// StdLogger is kept for backwards compatibility
var StdLogger *log.Logger

func init() {
	StdLogger = log.New(os.Stdout, "Zephyrix: ", log.LstdFlags)
	Logger = &zephyrixLogger{
		level:   LevelInfo,
		outputs: []io.Writer{os.Stdout},
	}
}
