package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHomeRedirectsToSDKGuide(t *testing.T) {
	handler := HomeHandler()
	for _, method := range []string{http.MethodGet, http.MethodHead} {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest(method, "/", nil))
		if w.Code != http.StatusFound || w.Header().Get("Location") != "/docs/byok.html" {
			t.Errorf("%s: %d %q", method, w.Code, w.Header().Get("Location"))
		}
	}
	for path, tc := range map[string]struct {
		method string
		want   int
	}{
		"/not-a-route": {http.MethodGet, http.StatusNotFound},
		"/":            {http.MethodPost, http.StatusMethodNotAllowed},
	} {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest(tc.method, path, nil))
		if w.Code != tc.want {
			t.Errorf("%s %s: %d", tc.method, path, w.Code)
		}
	}
}
