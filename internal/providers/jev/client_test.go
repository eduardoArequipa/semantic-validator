package jev

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClientEvaluate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/systemone" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Fatalf("missing authorization header")
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"model":"jev-latest"`) || !strings.Contains(string(body), `"state":"Jorge Eduardo"`) {
			t.Fatalf("unexpected body: %s", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"answers":{"result":{"choice":"true","probabilities":{"true":0.98,"false":0.02}}}}`))
	}))
	defer server.Close()

	client := NewClient(server.Client(), server.URL, "test-token", "jev-latest")
	result, err := client.Evaluate(context.Background(), "Jorge Eduardo", "¿Es un nombre?")
	if err != nil {
		t.Fatal(err)
	}
	if !result.Value || result.Confidence != .98 {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestClientRejectsInvalidResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"answers":{"result":{"choice":"true"}}}`))
	}))
	defer server.Close()

	client := NewClient(server.Client(), server.URL, "test-token", "jev-latest")
	if _, err := client.Evaluate(context.Background(), "x", "q"); err == nil {
		t.Fatal("expected invalid response error")
	}
}
