package logger

import (
	"bytes"
	"encoding/json"
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
func TestLevelStrings(t *testing.T) {
	tests := []struct {
		level Level
		want  string
	}{
		{TRACE, "TRACE"},
		{DEBUG, "DEBUG"},
		{INFO, "INFO"},
		{WARN, "WARN"},
		{ERROR, "ERROR"},
		{OFF, "OFF"},
		{Level(99), "UNKNOWN"},
	}
	for _, tt := range tests {
		if got := tt.level.String(); got != tt.want {
			t.Errorf("Level(%d).String() = %q, want %q", tt.level, got, tt.want)
		}
	}
}

func TestLevelColors(t *testing.T) {
	tests := []struct {
		level Level
		want  string
	}{
		{TRACE, "\033[35m"},
		{DEBUG, "\033[36m"},
		{INFO, "\033[32m"},
		{WARN, "\033[33m"},
		{ERROR, "\033[31m"},
		{Level(99), "\033[0m"},
	}
	for _, tt := range tests {
		if got := tt.level.Color(); got != tt.want {
			t.Errorf("Level(%d).Color() = %q, want %q", tt.level, got, tt.want)
		}
	}
}

func TestLoggerMethods(t *testing.T) {
	var buf bytes.Buffer
	l := New(DEBUG)
	l.Output = &buf
	l.Color = false

	l.Debug("d")
	if !strings.Contains(buf.String(), "[DEBUG] d") {
		t.Errorf("Debug() failed, got %q", buf.String())
	}
	buf.Reset()

	l.Warn("w")
	if !strings.Contains(buf.String(), "[WARN] w") {
		t.Errorf("Warn() failed, got %q", buf.String())
	}
	buf.Reset()
}

func TestGlobalLoggers(t *testing.T) {
	var buf bytes.Buffer
	oldOutput := Default.Output
	Default.Output = &buf
	oldLevel := Default.GetLevel()
	Default.SetLevel(DEBUG)
	Default.Color = false
	defer func() { 
		Default.Output = oldOutput 
		Default.SetLevel(oldLevel)
	}()

	Debug("dg")
	if !strings.Contains(buf.String(), "[DEBUG] dg") {
		t.Errorf("Global Debug() failed, got %q", buf.String())
	}
	buf.Reset()

	Info("ig")
	if !strings.Contains(buf.String(), "[INFO] ig") {
		t.Errorf("Global Info() failed, got %q", buf.String())
	}
	buf.Reset()

	Warn("wg")
	if !strings.Contains(buf.String(), "[WARN] wg") {
		t.Errorf("Global Warn() failed, got %q", buf.String())
	}
	buf.Reset()

	Error("eg")
	if !strings.Contains(buf.String(), "[ERROR] eg") {
		t.Errorf("Global Error() failed, got %q", buf.String())
	}
}

func TestLoggerJSONOutput(t *testing.T) {
	var buf bytes.Buffer
	l := New(INFO)
	l.Output = &buf
	l.SetJSON(true)

	l.Info("json message")
	var entry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to unmarshal JSON output: %v", err)
	}

	if entry["severity"] != "INFO" {
		t.Errorf("expected severity INFO, got %v", entry["severity"])
	}
	if entry["message"] != "json message" {
		t.Errorf("expected message 'json message', got %v", entry["message"])
	}
	if _, ok := entry["timestamp"]; !ok {
		t.Errorf("expected timestamp field, got %v", entry)
	}

	buf.Reset()
	l.Warn("warning message")
	json.Unmarshal(buf.Bytes(), &entry)
	if entry["severity"] != "WARNING" {
		t.Errorf("expected severity WARNING for Warn level, got %v", entry["severity"])
	}
}
