package logger

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogger_Levels(t *testing.T) {
	oldLevel := Default.GetLevel()
	defer SetLevel(oldLevel)

	SetLevel(LevelInfo)
	assert.Equal(t, LevelInfo, Default.GetLevel())
	assert.True(t, Default.LevelEnabled(LevelInfo))
	assert.True(t, Default.LevelEnabled(LevelError))
	assert.False(t, Default.LevelEnabled(LevelDebug))
}

func TestLogger_ParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected Level
	}{
		{"trace", LevelTrace},
		{"DEBUG", LevelDebug},
		{"Info", LevelInfo},
		{"WARN", LevelWarn},
		{"error", LevelError},
		{"OFF", LevelOff},
		{"invalid", LevelInfo}, // Default to Info
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, ParseLevel(tt.input))
	}
}

func TestLogger_ContextLogging(t *testing.T) {
	// These tests primarily verify that the functions don't panic
	// and execute correctly with context.
	ctx := context.Background()

	TraceCtx(ctx, "Trace message")
	DebugCtx(ctx, "Debug message")
	InfoCtx(ctx, "Info message")
	WarnCtx(ctx, "Warn message")
	ErrorCtx(ctx, "Error message")
}
