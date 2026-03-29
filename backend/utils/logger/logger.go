package logger

import (
	"fmt"
	"io"
	"os"
	"time"
)

// Level represents the severity of a log message.
type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

func (l Level) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Color returns the ANSI color code for the level.
func (l Level) Color() string {
	switch l {
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

// Logger provides a simple leveled logging system.
type Logger struct {
	Level  Level
	Output io.Writer
	Color  bool
}

// New creates a new Logger with the specified level.
func New(level Level) *Logger {
	return &Logger{
		Level:  level,
		Output: os.Stdout,
		Color:  true,
	}
}

func (l *Logger) log(level Level, msg string, args ...interface{}) {
	if level < l.Level {
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

// Debug logs a debug message using the default logger.
func Debug(msg string, args ...interface{}) { Default.Debug(msg, args...) }

// Info logs an informational message using the default logger.
func Info(msg string, args ...interface{}) { Default.Info(msg, args...) }

// Warn logs a warning message using the default logger.
func Warn(msg string, args ...interface{}) { Default.Warn(msg, args...) }

// Error logs an error message using the default logger.
func Error(msg string, args ...interface{}) { Default.Error(msg, args...) }
