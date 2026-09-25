package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/eduardoArequipa/semantic-validator/internal/config"
)

func TestDocsOnlyHandler(t *testing.T) {
	handler, err := buildHandler(config.Config{DocsOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/v1", "/v1/validate", "/v1/check", "/v1/validate/batch", "/v1/anything"} {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{"value":"test"}`)))
		if w.Code != http.StatusServiceUnavailable || !strings.Contains(w.Body.String(), "service_paused") {
			t.Errorf("%s: got %d %s", path, w.Code, w.Body.String())
		}
	}
	for _, path := range []string{"/docs/byok.html", "/docs/byok-en.html", "/openapi.yaml"} {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		if w.Code != http.StatusOK {
			t.Errorf("%s: got %d", path, w.Code)
		}
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
	if w.Code != http.StatusFound || w.Header().Get("Location") != "/docs/byok.html" {
		t.Fatalf("home redirect: %d %q", w.Code, w.Header().Get("Location"))
	}
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/health", nil))
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"mode":"docs_only"`) {
		t.Fatalf("health: %d %s", w.Code, w.Body.String())
	}
}
