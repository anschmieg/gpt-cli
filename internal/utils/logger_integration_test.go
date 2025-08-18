//go:build integration
// +build integration

package utils

import (
    "bytes"
    "log"
    "testing"
    "github.com/stretchr/testify/assert"
)

// Simple integration-style check that the logger can be swapped to a buffer and
// used without panics. Detailed concurrency tests live in unit.
func TestLogger_Integration(t *testing.T) {
    var buf bytes.Buffer
    l := &Logger{verbose: true, logger: log.New(&buf, "[DEBUG] ", log.LstdFlags)}
    l.Debug("x")
    l.Errorf("y %d", 1)
    out := buf.String()
    assert.Contains(t, out, "x")
    assert.Contains(t, out, "y 1")
}

