package utils

import (
	"log"
	"os"
)

// Logger provides logging functionality
type Logger struct {
	verbose bool
	logger  *log.Logger
}

// NewLogger creates a new logger instance
func NewLogger(verbose bool) *Logger {
	return &Logger{
		verbose: verbose,
		logger:  log.New(os.Stderr, "[DEBUG] ", log.LstdFlags),
	}
}

// Debug logs a debug message if verbose mode is enabled
func (l *Logger) Debug(v ...interface{}) {
	if l.verbose {
		l.logger.Println(v...)
	}
}

// Debugf logs a formatted debug message if verbose mode is enabled
func (l *Logger) Debugf(format string, v ...interface{}) {
	if l.verbose {
		l.logger.Printf(format, v...)
	}
}

// Error logs an error message
func (l *Logger) Error(v ...interface{}) {
	l.logger.Println(v...)
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.logger.Printf(format, v...)
}

// IsVerbose returns whether verbose logging is enabled
func (l *Logger) IsVerbose() bool {
	return l.verbose
}
