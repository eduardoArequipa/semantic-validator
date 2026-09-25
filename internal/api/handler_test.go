package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/eduardoArequipa/semantic-validator/internal/providers"
	"github.com/eduardoArequipa/semantic-validator/internal/rules"
	"github.com/eduardoArequipa/semantic-validator/internal/validation"
)

type handlerProvider struct{}

func (handlerProvider) Evaluate(context.Context, string, string) (providers.Result, error) {
	return providers.Result{Value: true, Confidence: .98}, nil
}

func TestValidateHandler(t *testing.T) {
	service := validation.NewService(rules.DefaultRegistry(), handlerProvider{})
	handler := NewValidateHandler(service)
	request := httptest.NewRequest(http.MethodPost, "/v1/validate", strings.NewReader(`{"rule":"person_name","value":"Jorge Eduardo"}`))
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	var result validation.Result
	if err := json.Unmarshal(recorder.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Valid == nil || !*result.Valid || result.Status != "valid" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestValidateHandlerRejectsUnknownFields(t *testing.T) {
	service := validation.NewService(rules.DefaultRegistry(), handlerProvider{})
	handler := NewValidateHandler(service)
	request := httptest.NewRequest(http.MethodPost, "/v1/validate", strings.NewReader(`{"rule":"person_name","value":"x","extra":true}`))
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestCheckHandler(t *testing.T) {
	service := validation.NewService(rules.DefaultRegistry(), handlerProvider{})
	handler := NewCheckHandler(service)
	request := httptest.NewRequest(http.MethodPost, "/v1/check", strings.NewReader(`{"value":"Necesito devolver el producto","question":"¿El cliente solicita una devolución?"}`))
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	var result validation.Result
	if err := json.Unmarshal(recorder.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Valid == nil || !*result.Valid || result.Status != "valid" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestBatchHandler(t *testing.T) {
	service := validation.NewService(rules.DefaultRegistry(), handlerProvider{})
	handler := NewBatchHandler(service)
	request := httptest.NewRequest(http.MethodPost, "/v1/validate/batch", strings.NewReader(`{"items":[{"id":"name","rule":"person_name","value":"Jorge Eduardo"},{"id":"other","rule":"person_name","value":"Ana"}]}`))
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	var response struct {
		Items []validation.BatchResult `json:"items"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if len(response.Items) != 2 || response.Items[0].ID != "name" || response.Items[1].ID != "other" {
		t.Fatalf("unexpected response: %+v", response)
	}
}

func TestOpenAPIHandler(t *testing.T) {
	recorder := httptest.NewRecorder()
	OpenAPIHandler(recorder, httptest.NewRequest(http.MethodGet, "/openapi.yaml", nil))
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), "openapi: 3.1.0") {
		t.Fatalf("unexpected OpenAPI response: status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}
