package render

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestSuccess(t *testing.T) {
	rr := httptest.NewRecorder()
	data := map[string]string{"message": "success"}

	Success(rr, data)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if rr.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected content-type application/json, got %s", rr.Header().Get("Content-Type"))
	}

	var res map[string]string
	err := json.NewDecoder(rr.Body).Decode(&res)
	if err != nil {
		t.Errorf("failed to decode body: %v", err)
	}
	if !reflect.DeepEqual(data, res) {
		t.Errorf("expected data %v, got %v", data, res)
	}
}

func TestError(t *testing.T) {
	rr := httptest.NewRecorder()
	msg := "something went wrong"
	code := http.StatusBadRequest

	Error(rr, msg, code)

	if rr.Code != code {
		t.Errorf("expected status %d, got %d", code, rr.Code)
	}
	if rr.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected content-type application/json, got %s", rr.Header().Get("Content-Type"))
	}

	var res map[string]string
	err := json.NewDecoder(rr.Body).Decode(&res)
	if err != nil {
		t.Errorf("failed to decode body: %v", err)
	}
	if res["error"] != msg {
		t.Errorf("expected error message %s, got %s", msg, res["error"])
	}
}

func TestJSON(t *testing.T) {
	rr := httptest.NewRecorder()
	data := map[string]interface{}{"foo": "bar", "baz": 123.0}
	code := http.StatusCreated

	JSON(rr, code, data)

	if rr.Code != code {
		t.Errorf("expected status %d, got %d", code, rr.Code)
	}
	if rr.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected content-type application/json, got %s", rr.Header().Get("Content-Type"))
	}

	var res map[string]interface{}
	err := json.NewDecoder(rr.Body).Decode(&res)
	if err != nil {
		t.Errorf("failed to decode body: %v", err)
	}
	if !reflect.DeepEqual(data, res) {
		t.Errorf("expected data %v, got %v", data, res)
	}
}
