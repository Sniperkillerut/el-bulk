package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type ctxKey string

const (
	traceKey ctxKey = "traceID"
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
	JSON      bool
	ProjectID string
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

// SetJSON enables or disables JSON logging mode.
func (l *Logger) SetJSON(enabled bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.JSON = enabled
}

// AutoDetectGCP automatically configures the logger for GCP environments.
func (l *Logger) AutoDetectGCP() {
	// Detect project ID from environment
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")

	// Detect Cloud Run or App Engine environment
	if os.Getenv("K_SERVICE") != "" || projectID != "" {
		l.mu.Lock()
		l.JSON = true
		l.Color = false
		l.ProjectID = projectID
		if l.level < INFO {
			l.level = INFO
		}
		l.mu.Unlock()
		// We don't log here because we might not be initialized yet
	}
}

func (l *Logger) log(ctx context.Context, level Level, msg string, args ...interface{}) {
	l.mu.RLock()
	currentLevel := l.level
	isJSON := l.JSON
	l.mu.RUnlock()

	if level < currentLevel || currentLevel == OFF {
		return
	}

	content := fmt.Sprintf(msg, args...)

	if isJSON {
		// GCP Severity Mapping
		severity := "INFO"
		switch level {
		case TRACE, DEBUG:
			severity = "DEBUG"
		case INFO:
			severity = "INFO"
		case WARN:
			severity = "WARNING"
		case ERROR:
			severity = "ERROR"
		}

		// Ensure the content is a single line for JSON (Marshal will escape \n)
		entry := map[string]interface{}{
			"severity":  severity,
			"message":   content,
			"timestamp": time.Now().Format(time.RFC3339),
		}

		// Add Trace ID if available in context
		if ctx != nil {
			if traceID, ok := ctx.Value(traceKey).(string); ok && traceID != "" {
				if l.ProjectID != "" {
					entry["logging.googleapis.com/trace"] = fmt.Sprintf("projects/%s/traces/%s", l.ProjectID, traceID)
				} else {
					// Minimal trace without project if unknown
					entry["logging.googleapis.com/trace"] = traceID
				}
			}
		}

		jsonBytes, _ := json.Marshal(entry)
		fmt.Fprintln(l.Output, string(jsonBytes))
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	levelStr := level.String()

	// If we are on GCP but JSON is off for some reason, we MUST escape newlines
	// to prevent fracturing.
	if os.Getenv("K_SERVICE") != "" {
		content = strings.ReplaceAll(content, "\n", " | ")
	}

	if l.Color {
		levelStr = level.Color() + levelStr + "\033[0m"
	}

	fmt.Fprintf(l.Output, "[%s] [%s] %s\n", timestamp, levelStr, content)
}

// Trace logs a trace message.
func (l *Logger) Trace(msg string, args ...interface{}) { l.log(context.Background(), TRACE, msg, args...) }

// Debug logs a debug message.
func (l *Logger) Debug(msg string, args ...interface{}) { l.log(context.Background(), DEBUG, msg, args...) }

// Info logs an informational message.
func (l *Logger) Info(msg string, args ...interface{}) { l.log(context.Background(), INFO, msg, args...) }

// Warn logs a warning message.
func (l *Logger) Warn(msg string, args ...interface{}) { l.log(context.Background(), WARN, msg, args...) }

// Error logs an error message.
func (l *Logger) Error(msg string, args ...interface{}) { l.log(context.Background(), ERROR, msg, args...) }

// TraceCtx logs a trace message with context.
func (l *Logger) TraceCtx(ctx context.Context, msg string, args ...interface{}) { l.log(ctx, TRACE, msg, args...) }

// DebugCtx logs a debug message with context.
func (l *Logger) DebugCtx(ctx context.Context, msg string, args ...interface{}) { l.log(ctx, DEBUG, msg, args...) }

// InfoCtx logs an informational message with context.
func (l *Logger) InfoCtx(ctx context.Context, msg string, args ...interface{}) { l.log(ctx, INFO, msg, args...) }

// WarnCtx logs a warning message with context.
func (l *Logger) WarnCtx(ctx context.Context, msg string, args ...interface{}) { l.log(ctx, WARN, msg, args...) }

// ErrorCtx logs an error message with context.
func (l *Logger) ErrorCtx(ctx context.Context, msg string, args ...interface{}) { l.log(ctx, ERROR, msg, args...) }

// Default is the global default logger instance.
var Default = New(INFO)

// SetLevel sets the default logger level.
func SetLevel(level Level) { Default.SetLevel(level) }

// SetJSON sets the default logger JSON mode.
func SetJSON(enabled bool) { Default.SetJSON(enabled) }

// AutoDetectGCP configures the default logger for GCP.
func AutoDetectGCP() { Default.AutoDetectGCP() }

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

// TraceCtx logs a trace message with context using the default logger.
func TraceCtx(ctx context.Context, msg string, args ...interface{}) { Default.TraceCtx(ctx, msg, args...) }

// DebugCtx logs a debug message with context using the default logger.
func DebugCtx(ctx context.Context, msg string, args ...interface{}) { Default.DebugCtx(ctx, msg, args...) }

// InfoCtx logs an informational message with context using the default logger.
func InfoCtx(ctx context.Context, msg string, args ...interface{}) { Default.InfoCtx(ctx, msg, args...) }

// WarnCtx logs a warning message with context using the default logger.
func WarnCtx(ctx context.Context, msg string, args ...interface{}) { Default.WarnCtx(ctx, msg, args...) }

// ErrorCtx logs an error message with context using the default logger.
func ErrorCtx(ctx context.Context, msg string, args ...interface{}) { Default.ErrorCtx(ctx, msg, args...) }

// RequestLogger is a middleware that logs the start and end of each request.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		ctx := r.Context()
		// Extract GCP Trace ID from header: X-Cloud-Trace-Context: TRACE_ID/SPAN_ID;o=TRACE_TRUE
		traceHeader := r.Header.Get("X-Cloud-Trace-Context")
		if traceHeader != "" {
			parts := strings.Split(traceHeader, "/")
			if len(parts) > 0 {
				ctx = context.WithValue(ctx, traceKey, parts[0])
			}
		}

		// Create a custom response writer to capture the status code
		ww := &statusResponseWriter{ResponseWriter: w, status: http.StatusOK}

		TraceCtx(ctx, "Request started: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

		next.ServeHTTP(ww, r.WithContext(ctx))

		duration := time.Since(start)

		// Log detailed info at DEBUG, summary at INFO
		detailedMsg := fmt.Sprintf("%d | %s | %s %s", ww.status, duration.String(), r.Method, r.URL.Path)
		summaryMsg := fmt.Sprintf("%s %s | %d", r.Method, r.URL.Path, ww.status)

		if ww.status >= 500 {
			ErrorCtx(ctx, "%s", detailedMsg)
		} else if ww.status >= 400 {
			WarnCtx(ctx, "%s", detailedMsg)
		} else {
			InfoCtx(ctx, "%s", summaryMsg)
			DebugCtx(ctx, "%s", detailedMsg)
		}
	})
}

// Recoverer is a middleware that recovers from panics and logs them through the structured logger.
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				if rvr == http.ErrAbortHandler {
					panic(rvr)
				}

				// Standard panic logging with stack trace
				ErrorCtx(r.Context(), "PANIC RECOVERED: %v", rvr)

				if r.Header.Get("Connection") != "Upgrade" {
					w.WriteHeader(http.StatusInternalServerError)
				}
			}
		}()

		next.ServeHTTP(w, r)
	})
}

type statusResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusResponseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}
