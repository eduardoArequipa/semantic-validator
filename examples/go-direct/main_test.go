package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	validator "github.com/eduardoArequipa/semantic-validator/sdk/go"
)

func TestRunWithoutRealKey(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.URL.Path != "/v1/systemone" || r.Header.Get("Authorization") != "Bearer test-only" {
			t.Errorf("unexpected provider request: %s", r.URL.Path)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read request: %v", err)
			return
		}
		confidence := 0.62
		if strings.Contains(string(body), "nombre de una persona") {
			confidence = 0.96
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"answers":{"result":{"choice":"true","confidence":%v}}}`, confidence)
	}))
	defer server.Close()

	client, err := validator.NewDirectClientWithOptions("test-only", validator.DirectOptions{
		BaseURL: server.URL, HTTPClient: server.Client(),
	})
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err := run(context.Background(), client, &output); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 2 || !strings.Contains(output.String(), "nombre: true") ||
		!strings.Contains(output.String(), "reclamo: requiere revisión") {
		t.Fatalf("unexpected demo output: calls=%d output=%q", calls.Load(), output.String())
	}
}

func TestProviderErrorIsNotFalse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "simulated failure", http.StatusServiceUnavailable)
	}))
	defer server.Close()
	client, err := validator.NewDirectClientWithOptions("test-only", validator.DirectOptions{
		BaseURL: server.URL, HTTPClient: server.Client(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := run(context.Background(), client, io.Discard); err == nil {
		t.Fatal("provider failure must return an error")
	}
}
