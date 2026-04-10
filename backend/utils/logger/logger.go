package logger

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

// Level represents the severity of a log message.
type Level int

const (
	TRACE Level = iota
	DEBUG
	INFO
	WARN
	ERROR
	OFF
)

func (l Level) String() string {
	switch l {
	case TRACE:
		return "TRACE"
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case OFF:
		return "OFF"
	default:
		return "UNKNOWN"
	}
}

// Color returns the ANSI color code for the level.
func (l Level) Color() string {
	switch l {
	case TRACE:
		return "\033[35m" // Magenta
	case DEBUG:
		return "\033[36m" // Cyan
	case INFO:
		return "\033[32m" // Green
	case WARN:
		return "\033[33m" // Yellow
	case ERROR:
		return "\033[31m" // Red
	default:
		return "\033[0m"
	}
}

// ParseLevel converts a string to a Level.
func ParseLevel(s string) Level {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "TRACE":
		return TRACE
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN":
		return WARN
	case "ERROR":
		return ERROR
	case "OFF":
		return OFF
	default:
		return INFO
	}
}

// Logger provides a simple leveled logging system.
type Logger struct {
	mu     sync.RWMutex
	level  Level
	Output io.Writer
	Color  bool
}

// New creates a new Logger with the specified level.
func New(level Level) *Logger {
	return &Logger{
		level:  level,
		Output: os.Stdout,
		Color:  true,
	}
}

// SetLevel sets the logger level thread-safely.
func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// GetLevel gets the logger level thread-safely.
func (l *Logger) GetLevel() Level {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.level
}

func (l *Logger) log(level Level, msg string, args ...interface{}) {
	l.mu.RLock()
	currentLevel := l.level
	l.mu.RUnlock()

	if level < currentLevel || currentLevel == OFF {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	levelStr := level.String()

	if l.Color {
		levelStr = level.Color() + levelStr + "\033[0m"
	}

	content := fmt.Sprintf(msg, args...)
	fmt.Fprintf(l.Output, "[%s] [%s] %s\n", timestamp, levelStr, content)
}

// Trace logs a trace message.
func (l *Logger) Trace(msg string, args ...interface{}) { l.log(TRACE, msg, args...) }

// Debug logs a debug message.
func (l *Logger) Debug(msg string, args ...interface{}) { l.log(DEBUG, msg, args...) }

// Info logs an informational message.
func (l *Logger) Info(msg string, args ...interface{}) { l.log(INFO, msg, args...) }

// Warn logs a warning message.
func (l *Logger) Warn(msg string, args ...interface{}) { l.log(WARN, msg, args...) }

// Error logs an error message.
func (l *Logger) Error(msg string, args ...interface{}) { l.log(ERROR, msg, args...) }

// Default is the global default logger instance.
var Default = New(INFO)

// SetLevel sets the default logger level.
func SetLevel(level Level) { Default.SetLevel(level) }

// Trace logs a trace message using the default logger.
func Trace(msg string, args ...interface{}) { Default.Trace(msg, args...) }

// Debug logs a debug message using the default logger.
func Debug(msg string, args ...interface{}) { Default.Debug(msg, args...) }

// Info logs an informational message using the default logger.
func Info(msg string, args ...interface{}) { Default.Info(msg, args...) }

// Warn logs a warning message using the default logger.
func Warn(msg string, args ...interface{}) { Default.Warn(msg, args...) }

// Error logs an error message using the default logger.
func Error(msg string, args ...interface{}) { Default.Error(msg, args...) }
