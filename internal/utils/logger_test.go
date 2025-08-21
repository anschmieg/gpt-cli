package utils

import (
	"bytes"
	"log"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name    string
		verbose bool
	}{
		{
			name:    "verbose enabled",
			verbose: true,
		},
		{
			name:    "verbose disabled",
			verbose: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLogger(tt.verbose)
			
			assert.NotNil(t, logger)
			assert.Equal(t, tt.verbose, logger.verbose)
			assert.NotNil(t, logger.logger)
			assert.Equal(t, tt.verbose, logger.IsVerbose())
		})
	}
}

func TestLoggerDebug(t *testing.T) {
	tests := []struct {
		name          string
		verbose       bool
		expectOutput  bool
	}{
		{
			name:         "verbose enabled - should log",
			verbose:      true,
			expectOutput: true,
		},
		{
			name:         "verbose disabled - should not log",
			verbose:      false,
			expectOutput: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stderr
			var buf bytes.Buffer
			logger := &Logger{
				verbose: tt.verbose,
				logger:  log.New(&buf, "[DEBUG] ", log.LstdFlags),
			}
			
			testMessage := "test debug message"
			logger.Debug(testMessage)
			
			output := buf.String()
			if tt.expectOutput {
				assert.Contains(t, output, testMessage)
				assert.Contains(t, output, "[DEBUG]")
			} else {
				assert.Empty(t, output)
			}
		})
	}
}

func TestLoggerDebugf(t *testing.T) {
	tests := []struct {
		name          string
		verbose       bool
		expectOutput  bool
	}{
		{
			name:         "verbose enabled - should log",
			verbose:      true,
			expectOutput: true,
		},
		{
			name:         "verbose disabled - should not log",
			verbose:      false,
			expectOutput: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stderr
			var buf bytes.Buffer
			logger := &Logger{
				verbose: tt.verbose,
				logger:  log.New(&buf, "[DEBUG] ", log.LstdFlags),
			}
			
			logger.Debugf("formatted message: %s %d", "test", 42)
			
			output := buf.String()
			if tt.expectOutput {
				assert.Contains(t, output, "formatted message: test 42")
				assert.Contains(t, output, "[DEBUG]")
			} else {
				assert.Empty(t, output)
			}
		})
	}
}

func TestLoggerError(t *testing.T) {
	// Capture stderr
	var buf bytes.Buffer
	logger := &Logger{
		verbose: false, // Error should log regardless of verbose setting
		logger:  log.New(&buf, "[DEBUG] ", log.LstdFlags),
	}
	
	testMessage := "test error message"
	logger.Error(testMessage)
	
	output := buf.String()
	assert.Contains(t, output, testMessage)
	assert.Contains(t, output, "[DEBUG]") // Uses same prefix
}

func TestLoggerErrorf(t *testing.T) {
	// Capture stderr
	var buf bytes.Buffer
	logger := &Logger{
		verbose: false, // Error should log regardless of verbose setting
		logger:  log.New(&buf, "[DEBUG] ", log.LstdFlags),
	}
	
	logger.Errorf("formatted error: %s %d", "test", 500)
	
	output := buf.String()
	assert.Contains(t, output, "formatted error: test 500")
	assert.Contains(t, output, "[DEBUG]") // Uses same prefix
}

func TestLoggerIsVerbose(t *testing.T) {
	verboseLogger := NewLogger(true)
	assert.True(t, verboseLogger.IsVerbose())
	
	quietLogger := NewLogger(false)
	assert.False(t, quietLogger.IsVerbose())
}

func TestLoggerWithRealStderr(t *testing.T) {
	// Test with actual stderr to ensure NewLogger works correctly
	logger := NewLogger(true)
	
	// This should work without panicking
	logger.Debug("test message")
	logger.Debugf("formatted %s", "message")
	logger.Error("error message")
	logger.Errorf("formatted %s", "error")
	
	// No assertions here since we can't easily capture stderr in this test,
	// but this ensures the logger doesn't crash
}

func TestLoggerConcurrency(t *testing.T) {
	// Test that logger is safe for concurrent use
	var buf bytes.Buffer
	logger := &Logger{
		verbose: true,
		logger:  log.New(&buf, "[DEBUG] ", log.LstdFlags),
	}
	
	// Start multiple goroutines writing to the logger
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func(id int) {
			logger.Debugf("goroutine %d message", id)
			logger.Error("error from goroutine", id)
			done <- true
		}(i)
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
	
	output := buf.String()
	// Should contain messages from all goroutines
	assert.Contains(t, output, "goroutine")
	assert.Contains(t, output, "error from goroutine")
	
	// Count number of log entries (rough test)
	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.GreaterOrEqual(t, len(lines), 10) // At least 10 messages
}