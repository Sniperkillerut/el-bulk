package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogHandler_Receive(t *testing.T) {
	h := NewLogHandler()

	t.Run("Success Info", func(t *testing.T) {
		log := FrontendLog{
			Level:   "info",
			Message: "Test message",
			Context: map[string]interface{}{"user_id": "123"},
		}
		body, _ := json.Marshal(log)

		req, _ := http.NewRequest("POST", "/api/logs", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		h.Receive(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)
	})

	t.Run("Success Error", func(t *testing.T) {
		log := FrontendLog{
			Level:   "error",
			Message: "Critical error",
		}
		body, _ := json.Marshal(log)

		req, _ := http.NewRequest("POST", "/api/logs", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		h.Receive(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/logs", bytes.NewBuffer([]byte("{invalid}")))
		rr := httptest.NewRecorder()
		h.Receive(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("Unknown Level", func(t *testing.T) {
		log := FrontendLog{
			Level:   "custom",
			Message: "Custom trace",
		}
		body, _ := json.Marshal(log)

		req, _ := http.NewRequest("POST", "/api/logs", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		h.Receive(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)
	})
}
