package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/el-bulk/backend/utils/logger"
)

type FrontendLog struct {
	Level   string                 `json:"level"`
	Message string                 `json:"message"`
	Context map[string]interface{} `json:"context,omitempty"`
}

type LogHandler struct{}

func NewLogHandler() *LogHandler {
	return &LogHandler{}
}

func (h *LogHandler) Receive(w http.ResponseWriter, r *http.Request) {
	var log FrontendLog
	if err := json.NewDecoder(r.Body).Decode(&log); err != nil {
		http.Error(w, "Invalid log format", http.StatusBadRequest)
		return
	}

	// Format context for logging if it exists
	ctxStr := ""
	if len(log.Context) > 0 {
		ctxBytes, _ := json.Marshal(log.Context)
		ctxStr = " | context: " + string(ctxBytes)
	}

	prefix := "[frontend] "
	switch log.Level {
	case "trace":
		logger.TraceCtx(r.Context(), "%s%s%s", prefix, log.Message, ctxStr)
	case "debug":
		logger.DebugCtx(r.Context(), "%s%s%s", prefix, log.Message, ctxStr)
	case "info":
		logger.InfoCtx(r.Context(), "%s%s%s", prefix, log.Message, ctxStr)
	case "warn":
		logger.WarnCtx(r.Context(), "%s%s%s", prefix, log.Message, ctxStr)
	case "error":
		logger.ErrorCtx(r.Context(), "%s%s%s", prefix, log.Message, ctxStr)
	default:
		logger.InfoCtx(r.Context(), "%s[%s] %s%s", prefix, log.Level, log.Message, ctxStr)
	}

	w.WriteHeader(http.StatusNoContent)
}
