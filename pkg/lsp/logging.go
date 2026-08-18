package lsp

import (
	"fmt"
	"log"
	"strings"
)

// LogLevel controls which server events are written to the configured logger.
type LogLevel uint8

const (
	// LogLevelError writes only server and transport failures.
	LogLevelError LogLevel = iota
	// LogLevelInfo writes lifecycle events and failures.
	LogLevelInfo
	// LogLevelDebug also writes high-frequency events such as document changes.
	LogLevelDebug
)

// ParseLogLevel converts a command-line log level to its server value.
func ParseLogLevel(value string) (LogLevel, error) {
	switch strings.ToLower(value) {
	case "error":
		return LogLevelError, nil
	case "info":
		return LogLevelInfo, nil
	case "debug":
		return LogLevelDebug, nil
	default:
		return LogLevelInfo, fmt.Errorf("invalid log level %q (want error, info, or debug)", value)
	}
}

// Option configures a Server.
type Option func(*Server)

// WithLogger directs server diagnostics to logger.
func WithLogger(logger *log.Logger) Option {
	return func(server *Server) {
		if logger != nil {
			server.logger = logger
		}
	}
}

// WithLogLevel sets the minimum detail emitted by the server logger.
func WithLogLevel(level LogLevel) Option {
	return func(server *Server) {
		server.logLevel = level
	}
}

func (s *Server) logf(level LogLevel, format string, arguments ...any) {
	if level <= s.logLevel {
		s.logger.Printf(format, arguments...)
	}
}
