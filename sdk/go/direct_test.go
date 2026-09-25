package validator

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

type directTransport func(*http.Request) (*http.Response, error)

func (f directTransport) RoundTrip(request *http.Request) (*http.Response, error) { return f(request) }

func TestDirectClient(t *testing.T) {
	var calls int
	transport := directTransport(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.URL.String() != "https://api.typesafe.ai/v1/systemone" || r.Header.Get("Authorization") != "Bearer own-key" {
			t.Errorf("wrong endpoint or bearer header")
		}
		var body struct {
			Model     string `json:"model"`
			State     string `json:"state"`
			Questions map[string]struct {
				Type         string `json:"type"`
				Instructions string `json:"instructions"`
			} `json:"questions"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		if body.Model != "jev-latest" || body.State == "" || body.Questions["result"].Type != "choice" {
			t.Errorf("wrong Jev payload: %+v", body)
		}
		choice, confidence := "true", "0.93"
		if body.State == "No" {
			choice, confidence = "false", "0.98"
		}
		if body.State == "Maybe" {
			confidence = "0.63"
		}
		response := `{"answers":{"result":{"type":"choice","choice":"` + choice + `","confidence":` + confidence + `}}}`
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(response)), Header: make(http.Header)}, nil
	})
	client, err := NewDirectClientWithOptions("own-key", DirectOptions{HTTPClient: &http.Client{Transport: transport}})
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.Name(context.Background(), "Jorge Eduardo")
	if err != nil || result.Valid == nil || !*result.Valid || result.Status != StatusValid {
		t.Fatalf("unexpected result: %+v %v", result, err)
	}
	negative, err := client.Check(context.Background(), "No", "¿Compra?")
	if err != nil || negative.Valid == nil || *negative.Valid || negative.Status != StatusInvalid {
		t.Fatalf("unexpected negative result: %+v %v", negative, err)
	}
	uncertain, err := client.Check(context.Background(), "Maybe", "Question")
	if err != nil || uncertain.Valid != nil || uncertain.Status != StatusUncertain {
		t.Fatalf("unexpected uncertain result: %+v %v", uncertain, err)
	}
	batch, err := client.ValidateBatch(context.Background(), []BatchItem{
		{ID: "a", Rule: "person_name", Value: "Jorge"},
		{ID: "b", Rule: "missing", Value: "No"},
		{ID: "c", Rule: "person_name", Value: " "},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(batch) != 3 || batch[0].ID != "a" || batch[1].Error.Code != "unknown_rule" || batch[2].Error.Code != "invalid_request" {
		t.Fatalf("unexpected batch: %+v", batch)
	}
	if calls != 4 {
		t.Fatalf("expected 4 Jev calls, got %d", calls)
	}
}

func TestDirectInvalidResponseAndConfig(t *testing.T) {
	if _, err := NewDirectClient(""); err == nil {
		t.Fatal("missing key accepted")
	}
	if _, err := NewDirectClientWithOptions("key", DirectOptions{BaseURL: "http://example.com"}); err == nil {
		t.Fatal("insecure endpoint accepted")
	}
	transport := directTransport(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{}`)), Header: make(http.Header)}, nil
	})
	client, err := NewDirectClientWithOptions("key", DirectOptions{HTTPClient: &http.Client{Transport: transport}})
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Check(context.Background(), "text", "question")
	if directErr, ok := err.(*DirectError); !ok || directErr.Code != "invalid_response" {
		t.Fatalf("wrong error: %v", err)
	}
}
