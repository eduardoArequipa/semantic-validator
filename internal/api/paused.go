package api

import (
	"net/http"
)

// PausedHandler keeps documentation online without accepting validation traffic.
func PausedHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusServiceUnavailable)
	_, _ = w.Write([]byte(`{"error":{"code":"service_paused","message":"Hosted validation is paused. Run Semantic Validator with your own Jev key; see /docs/byok.html."}}`))
}

func DocsOnlyHealthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"ok","mode":"docs_only"}`))
}
