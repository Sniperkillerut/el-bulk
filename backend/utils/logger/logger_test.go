package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogger_Levels(t *testing.T) {
	oldLevel := Default.GetLevel()
	defer SetLevel(oldLevel)

	SetLevel(INFO)
	assert.Equal(t, INFO, Default.GetLevel())
	assert.True(t, Default.LevelEnabled(INFO))
	assert.True(t, Default.LevelEnabled(ERROR))
	assert.False(t, Default.LevelEnabled(DEBUG))
}

func TestLogger_ParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected Level
	}{
		{"trace", TRACE},
		{"DEBUG", DEBUG},
		{"Info", INFO},
		{"WARN", WARN},
		{"error", ERROR},
		{"OFF", OFF},
		{"invalid", INFO}, // Default to Info
	}

	for _, tt := range tests {
		got := ParseLevel(tt.input)
		if got != tt.expected {
			t.Errorf("ParseLevel(%q) = %d, want %d", tt.input, got, tt.expected)
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

func TestLoggerTraceID(t *testing.T) {
	var buf bytes.Buffer
	l := New(INFO)
	l.Output = &buf
	l.SetJSON(true)
	l.ProjectID = "test-project"

	ctx := context.WithValue(context.Background(), traceKey, "test-trace-id")
	l.InfoCtx(ctx, "trace message")

	var entry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to unmarshal JSON output: %v", err)
	}

	expectedTrace := "projects/test-project/traces/test-trace-id"
	if entry["logging.googleapis.com/trace"] != expectedTrace {
		t.Errorf("expected trace %q, got %q", expectedTrace, entry["logging.googleapis.com/trace"])
	}
}

func TestAutoDetectGCP(t *testing.T) {
	// Backup env
	oldProject := os.Getenv("GOOGLE_CLOUD_PROJECT")
	oldService := os.Getenv("K_SERVICE")
	defer func() {
		os.Setenv("GOOGLE_CLOUD_PROJECT", oldProject)
		os.Setenv("K_SERVICE", oldService)
	}()

	os.Setenv("GOOGLE_CLOUD_PROJECT", "my-project")
	os.Setenv("K_SERVICE", "my-service")

	l := New(DEBUG)
	l.AutoDetectGCP()

	if !l.JSON {
		t.Error("expected JSON to be true after AutoDetectGCP")
	}
	if l.ProjectID != "my-project" {
		t.Errorf("expected ProjectID 'my-project', got %q", l.ProjectID)
	}
}

func TestRequestLogger(t *testing.T) {
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

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	loggerHandler := RequestLogger(handler)

	req := httptest.NewRequest("GET", "/test-path", nil)
	rr := httptest.NewRecorder()

	loggerHandler.ServeHTTP(rr, req)

	output := buf.String()

	// Should contain INFO summary
	if !strings.Contains(output, "[INFO] GET /test-path | 200") {
		t.Errorf("expected INFO summary log, got %q", output)
	}

	// Should contain DEBUG detailed info
	if !strings.Contains(output, "[DEBUG] 200 |") || !strings.Contains(output, "GET /test-path") {
		t.Errorf("expected DEBUG detailed log, got %q", output)
	}
}
