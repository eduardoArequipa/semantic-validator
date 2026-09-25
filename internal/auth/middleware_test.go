package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddlewareRequiresBearerKey(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	handler := Middleware("local-dev", next)

	tests := []struct {
		name          string
		authorization string
		status        int
	}{
		{name: "missing", status: http.StatusUnauthorized},
		{name: "wrong", authorization: "Bearer wrong", status: http.StatusUnauthorized},
		{name: "valid", authorization: "Bearer local-dev", status: http.StatusNoContent},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/v1/validate", nil)
			if test.authorization != "" {
				request.Header.Set("Authorization", test.authorization)
			}
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, request)
			if recorder.Code != test.status {
				t.Fatalf("status=%d want=%d", recorder.Code, test.status)
			}
		})
	}
}
