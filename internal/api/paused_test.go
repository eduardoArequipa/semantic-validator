package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPausedHandler(t *testing.T) {
	w := httptest.NewRecorder()
	PausedHandler(w, httptest.NewRequest(http.MethodPost, "/v1/check", nil))
	if w.Code != http.StatusServiceUnavailable || w.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("paused response: %d, %v", w.Code, w.Header())
	}
	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil || body.Error.Code != "service_paused" {
		t.Fatalf("paused body: %s (%v)", w.Body.String(), err)
	}
}
