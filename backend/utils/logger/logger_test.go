package logger

import (
	"bytes"
	"strings"
	"testing"
)

func TestLoggerLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	l := New(INFO)
	l.Output = &buf
	l.Color = false

	l.Debug("this should not appear")
	if buf.Len() > 0 {
		t.Errorf("expected empty buffer for Debug log, got %q", buf.String())
	}

	l.Info("this should appear")
	if !strings.Contains(buf.String(), "[INFO] this should appear") {
		t.Errorf("expected Info log in buffer, got %q", buf.String())
	}
	buf.Reset()

	l.Error("this should also appear")
	if !strings.Contains(buf.String(), "[ERROR] this should also appear") {
		t.Errorf("expected Error log in buffer, got %q", buf.String())
	}
}

func TestLoggerColorOutput(t *testing.T) {
	var buf bytes.Buffer
	l := New(DEBUG)
	l.Output = &buf
	l.Color = true

	l.Info("colored info")
	// Level.Color() for INFO is \033[32m (Green)
	if !strings.Contains(buf.String(), "\033[32mINFO\033[0m") {
		t.Errorf("expected colored level in output, got %q", buf.String())
	}
}

func TestFormatting(t *testing.T) {
	var buf bytes.Buffer
	l := New(INFO)
	l.Output = &buf
	l.Color = false

	l.Info("hello %s", "world")
	if !strings.Contains(buf.String(), "hello world") {
		t.Errorf("expected formatted output, got %q", buf.String())
	}
}
