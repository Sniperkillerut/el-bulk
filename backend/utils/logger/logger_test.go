package logger

import (
	"context"
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
