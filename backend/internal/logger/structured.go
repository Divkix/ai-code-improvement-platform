// ABOUTME: Structured logging implementation with correlation ID support
// ABOUTME: Provides centralized logging configuration and request tracing capabilities
package logger

import (
	"io"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// StructuredLogger wraps logrus.Logger with additional functionality
type StructuredLogger struct {
	*logrus.Logger
}

// Config holds logging configuration
type Config struct {
	Level  string // debug, info, warn, error
	Format string // json, text
	Output string // stdout, stderr, file
}

// NewStructuredLogger creates a new structured logger with the given configuration
func NewStructuredLogger(cfg Config) *StructuredLogger {
	logger := logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// Set formatter
	switch strings.ToLower(cfg.Format) {
	case "json":
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
	case "text":
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	default:
		// Default to JSON for production
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}

	// Set output
	var output io.Writer
	switch strings.ToLower(cfg.Output) {
	case "stderr":
		output = os.Stderr
	case "stdout":
		output = os.Stdout
	default:
		output = os.Stdout
	}
	logger.SetOutput(output)

	return &StructuredLogger{Logger: logger}
}

// WithCorrelation creates a log entry with correlation ID and service context
func (l *StructuredLogger) WithCorrelation(correlationID string) *logrus.Entry {
	return l.WithFields(logrus.Fields{
		"correlation_id": correlationID,
		"service":        "github-analyzer",
	})
}

// WithRequest creates a log entry with request context information
func (l *StructuredLogger) WithRequest(correlationID, method, path, userAgent, clientIP string) *logrus.Entry {
	return l.WithFields(logrus.Fields{
		"correlation_id": correlationID,
		"service":        "github-analyzer",
		"http_method":    method,
		"http_path":      path,
		"user_agent":     userAgent,
		"client_ip":      clientIP,
	})
}

// WithError creates a log entry with error context
func (l *StructuredLogger) WithError(correlationID string, err error) *logrus.Entry {
	return l.WithFields(logrus.Fields{
		"correlation_id": correlationID,
		"service":        "github-analyzer",
		"error":          err.Error(),
	})
}

// WithUser creates a log entry with user context
func (l *StructuredLogger) WithUser(correlationID, userID, email string) *logrus.Entry {
	return l.WithFields(logrus.Fields{
		"correlation_id": correlationID,
		"service":        "github-analyzer",
		"user_id":        userID,
		"user_email":     email,
	})
}

// WithDatabase creates a log entry with database operation context
func (l *StructuredLogger) WithDatabase(correlationID, operation, collection string) *logrus.Entry {
	return l.WithFields(logrus.Fields{
		"correlation_id": correlationID,
		"service":        "github-analyzer",
		"db_operation":   operation,
		"db_collection":  collection,
	})
}